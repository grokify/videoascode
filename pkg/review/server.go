package review

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	//go:embed static/index.html
	pageHTML string
	//go:embed static/review.css
	pageCSS string
	//go:embed static/review.js
	pageJS string
	// speechJS is the standalone build of @plexusone/speech, copied in by
	// `make vendor-speech`. It powers the browser speech preview.
	//go:embed static/vendor/speech.js
	speechJS string
)

// DeckPrefix is the path the handler serves the deck's directory under.
const DeckPrefix = "/deck/"

// HandlerOptions configures NewHandler.
type HandlerOptions struct {
	// Load builds the model from the files on disk. It runs on every model
	// request, so edits to the deck or transcript show on reload.
	Load func() (*Model, error)
	// ReviewPath is the review file decisions are saved to.
	ReviewPath string
	// DeckDir is the deck's directory, served under DeckPrefix.
	DeckDir string
	// Reviewer is recorded with each decision; optional.
	Reviewer string
	// Now returns the current time; default time.Now.
	Now func() time.Time
}

// Decision is the body of a POST to /api/review. An empty Status clears the
// slide's decision.
type Decision struct {
	SlideID  string `json:"slideId"`
	Language string `json:"language"`
	Status   Status `json:"status"`
	Note     string `json:"note"`
}

// NewHandler serves the review page:
//
//	GET  /              the review page
//	GET  /api/model     the model, rebuilt from disk
//	POST /api/review    record a Decision; returns the rebuilt model
//	GET  /deck/...      the deck's directory
func NewHandler(opts HandlerOptions) http.Handler {
	if opts.Now == nil {
		opts.Now = time.Now
	}
	h := &handler{opts: opts}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.page)
	mux.HandleFunc("GET /api/model", h.model)
	mux.HandleFunc("POST /api/review", h.decide)
	mux.Handle("GET "+DeckPrefix, http.StripPrefix(DeckPrefix, noCache(http.FileServerFS(os.DirFS(opts.DeckDir)))))
	return mux
}

type handler struct {
	opts HandlerOptions
	mu   sync.Mutex // serializes review file writes
}

func (h *handler) page(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := writePage(w, "Transcript review", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handler) model(w http.ResponseWriter, _ *http.Request) {
	m, err := h.opts.Load()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, m)
}

func (h *handler) decide(w http.ResponseWriter, r *http.Request) {
	var d Decision
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&d); err != nil {
		http.Error(w, "invalid decision: "+err.Error(), http.StatusBadRequest)
		return
	}
	if d.Status != "" && d.Status != StatusApproved && d.Status != StatusChanges {
		http.Error(w, fmt.Sprintf("invalid status %q", d.Status), http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	m, err := h.opts.Load()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	n, ok := m.narration(d.SlideID, d.Language)
	if !ok {
		http.Error(w, fmt.Sprintf("no slide %q in language %q", d.SlideID, d.Language), http.StatusNotFound)
		return
	}
	f, err := LoadFile(h.opts.ReviewPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if d.Status == "" {
		f.Clear(d.SlideID, d.Language)
	} else {
		f.Set(d.SlideID, d.Language, Entry{
			Status:    d.Status,
			Note:      strings.TrimSpace(d.Note),
			TextHash:  n.Hash,
			Reviewer:  h.opts.Reviewer,
			UpdatedAt: h.opts.Now().UTC().Truncate(time.Second),
		})
	}
	if err := f.Save(h.opts.ReviewPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if m, err = h.opts.Load(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, m)
}

func (m *Model) narration(slideID, lang string) (Narration, bool) {
	for _, s := range m.Slides {
		if s.ID == slideID {
			n, ok := s.Langs[lang]
			return n, ok
		}
	}
	return Narration{}, false
}

// Export writes a read-only review page with the model embedded, for sharing
// or archiving. m.Frame.URL must point to the deck relative to where the page
// will be saved.
func Export(w io.Writer, m *Model) error {
	ro := *m
	ro.ReadOnly = true
	title := "Transcript review"
	if m.Title != "" {
		title = m.Title + " · " + title
	}
	return writePage(w, title, &ro)
}

// writePage renders the page, embedding m when it is not nil; otherwise the
// page fetches the model from /api/model.
func writePage(w io.Writer, title string, m *Model) error {
	data := []byte("null")
	if m != nil {
		var err error
		// json.Marshal escapes <, > and &, so the model cannot close the
		// script element it is embedded in.
		if data, err = json.Marshal(m); err != nil {
			return err
		}
	}
	page := strings.NewReplacer(
		"{{TITLE}}", htmlEscape(title),
		"{{CSS}}", pageCSS,
		"{{SPEECH}}", speechJS,
		"{{JS}}", pageJS,
		"{{MODEL}}", string(data),
	).Replace(pageHTML)
	_, err := io.WriteString(w, page)
	return err
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		h.ServeHTTP(w, r)
	})
}

func htmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
