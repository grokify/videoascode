package review

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

//go:embed static/index-decks.html
var decksHTML string

var decksTmpl = template.Must(template.New("decks").Funcs(template.FuncMap{
	"pct": func(n, total int) int {
		if total == 0 {
			return 0
		}
		return 100 * n / total
	},
}).Parse(decksHTML))

// DeckOptions is one deck in a multi-deck review.
type DeckOptions struct {
	// Slug names the deck in URLs (/d/<slug>/); it must be unique.
	Slug string
	// Name is shown in the deck list and switcher; default the slug.
	Name string
	HandlerOptions
}

var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

// NewMultiHandler serves several decks for review in one session:
//
//	GET /            a list of decks with their review progress
//	GET /d/<slug>/   one deck's review page (see NewHandler)
//
// Each deck's model lists every deck (Model.Decks), so its page can switch
// between them.
func NewMultiHandler(decks []DeckOptions) (http.Handler, error) {
	if len(decks) == 0 {
		return nil, fmt.Errorf("review: no decks")
	}
	seen := map[string]bool{}
	for _, d := range decks {
		if !slugRe.MatchString(d.Slug) {
			return nil, fmt.Errorf("review: invalid deck slug %q", d.Slug)
		}
		if seen[d.Slug] {
			return nil, fmt.Errorf("review: duplicate deck slug %q", d.Slug)
		}
		seen[d.Slug] = true
	}

	mux := http.NewServeMux()
	for _, d := range decks {
		opts := d.HandlerOptions
		load := opts.Load
		opts.Load = func() (*Model, error) {
			m, err := load()
			if err != nil {
				return nil, err
			}
			for _, o := range decks {
				m.Decks = append(m.Decks, DeckLink{
					Name:    deckName(o, nil),
					URL:     "../" + o.Slug + "/",
					Current: o.Slug == d.Slug,
				})
			}
			return m, nil
		}
		prefix := "/d/" + d.Slug
		mux.Handle(prefix+"/", http.StripPrefix(prefix, NewHandler(opts)))
	}
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		type row struct {
			Name, URL string
			Model     *Model
			Err       string
		}
		rows := make([]row, len(decks))
		for i, d := range decks {
			rows[i] = row{URL: "d/" + d.Slug + "/"}
			m, err := d.Load()
			if err != nil {
				rows[i].Name, rows[i].Err = deckName(d, nil), err.Error()
				continue
			}
			rows[i].Name, rows[i].Model = deckName(d, m), m
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		// pageCSS is the embedded stylesheet, never user input.
		css := template.CSS(pageCSS) //nolint:gosec // G203: trusted, embedded at build time
		if err := decksTmpl.Execute(w, map[string]any{"Decks": rows, "CSS": css}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	return mux, nil
}

func deckName(d DeckOptions, m *Model) string {
	switch {
	case d.Name != "":
		return d.Name
	case m != nil && m.Title != "":
		return m.Title
	}
	return strings.ReplaceAll(d.Slug, "-", " ")
}
