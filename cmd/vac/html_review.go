package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/grokify/videoascode/pkg/review"
	"github.com/spf13/cobra"
)

// Flags for 'vac html review'.
var (
	hrPort       int
	hrOpen       bool
	hrReviewPath string
	hrReviewer   string
	hrLanguages  []string
	hrExport     string
	hrNav        string
	hrWidth      int
	hrHeight     int
	hrSummary    bool
)

// Flag for 'vac html video'.
var hvRequireReview bool

var htmlReviewCmd = &cobra.Command{
	Use:   "review [deck.html ...]",
	Short: "Review a deck's narration slide by slide before sending it to TTS",
	Long: `Serve a review page that shows each slide next to its narration, so a person
can read the transcript before any of it is sent to a TTS provider.

For each slide and language the page shows:

  - the narration, sentence by sentence, with its pauses
  - the text TTS will receive, after pronunciation substitutions
  - words and estimated seconds
  - checks for text TTS tends to misread: symbols, addresses, numbers,
    acronyms with no pronunciation entry, very long sentences, and missing
    translations

The reviewer approves each slide or requests changes with a note. Decisions
are saved to the review file (voiceover/review.json next to the transcript,
by default), keyed by slide ID and language. A decision records a hash of the
text it was made on: editing a slide's narration, pause, or pronunciations
makes it stale, and the slide needs review again.

Name several decks (with --input and as arguments) to review them in one
session: the page lists every deck with its progress, and each deck's page
can switch to the others. Each deck keeps its own transcript and review file.

The page reads the deck and transcript from disk on every load, so edit the
deck and press R to see the change. Narration itself is edited in the deck
(or the translation in the transcript), not on the page.

Use 'vac html video --require-review' to refuse recording until every slide
being recorded is approved in every language being recorded.

With --export, write a read-only snapshot page instead of serving, for sharing
the transcript or archiving a review. With --summary, print review progress
and exit.`,
	Example: `  vac html review --input deck/index.html --open
  vac html review --input deck/index.html --pronunciations pronunciations.json --reviewer alex
  vac html review --input deck/index.html --lang en-US,fr-FR
  vac html review course/sections/*/index.html --open
  vac html review --input deck/index.html --export deck/transcript.html
  vac html review course/sections/*/index.html --summary`,
	RunE: runHTMLReview,
}

func init() {
	f := htmlReviewCmd.Flags()
	f.IntVar(&hrPort, "port", 8765, "Port to serve the review page on (0 for any free port)")
	f.BoolVar(&hrOpen, "open", false, "Open the review page in the default browser")
	f.StringVar(&hrReviewPath, "review-file", "", "Review decisions file (default: review.json next to the transcript; one deck only)")
	f.StringVar(&hrReviewer, "reviewer", "", "Name recorded with each decision (optional)")
	f.StringSliceVarP(&hrLanguages, "lang", "l", nil, "Languages to review (default: every language in the transcript)")
	f.StringVar(&hrExport, "export", "", "Write a read-only snapshot page to this file instead of serving (one deck only)")
	f.StringVar(&hrNav, "nav", review.NavHash, `How the page moves the deck between slides: "hash" (change the URL fragment) or "reload" (reload at each slide's fragment)`)
	f.IntVar(&hrWidth, "width", 1920, "Width the deck is laid out at; the page scales it to fit")
	f.IntVar(&hrHeight, "height", 1080, "Height the deck is laid out at")
	f.BoolVar(&hrSummary, "summary", false, "Print review progress per language and exit")

	htmlVideoCmd.Flags().BoolVar(&hvRequireReview, "require-review", false, "Refuse to record unless every slide being recorded is approved in 'vac html review' for every --lang")
}

// reviewFilePath is --review-file, defaulting to review.json next to the
// transcript.
func reviewFilePath(transcriptPath string) string {
	if hrReviewPath != "" {
		return hrReviewPath
	}
	return filepath.Join(filepath.Dir(transcriptPath), review.DefaultFile)
}

// reviewDecks returns the decks to review: --input, then the arguments.
// Flags that name one file are refused with more than one deck.
func reviewDecks(args []string) ([]string, error) {
	var decks []string
	for _, d := range append([]string{hdInput}, args...) {
		if d != "" && !slices.Contains(decks, d) {
			decks = append(decks, d)
		}
	}
	if len(decks) == 0 {
		return nil, fmt.Errorf("name a deck with --input or as an argument")
	}
	if len(decks) > 1 {
		for flag, v := range map[string]string{"--transcript": hdTranscriptPath, "--review-file": hrReviewPath, "--export": hrExport} {
			if v != "" {
				return nil, fmt.Errorf("%s names one file; use it with one deck", flag)
			}
		}
	}
	return decks, nil
}

// buildReviewModel reads a deck, its transcript, pronunciations, and review
// file from disk and describes them for review.
func buildReviewModel(input, frameURL string, langs []string, readOnly bool) (*review.Model, *htmlDeckInputs, error) {
	in, err := loadHTMLDeckFile(input)
	if err != nil {
		return nil, nil, err
	}
	rf, err := review.LoadFile(reviewFilePath(in.transcriptPath))
	if err != nil {
		return nil, nil, err
	}
	m, err := review.Build(review.Input{
		Deck:           in.deck,
		Transcript:     in.transcript,
		Pronunciations: in.pronunciations,
		Review:         rf,
		Languages:      langs,
		Frame: review.Frame{
			URL:    frameURL,
			Query:  hdQuery,
			Width:  hrWidth,
			Height: hrHeight,
			Nav:    hrNav,
		},
		StartFragment: hdStartFragment,
		ReadOnly:      readOnly,
		Generated:     time.Now().UTC().Truncate(time.Second),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", input, err)
	}
	return m, in, nil
}

func runHTMLReview(cmd *cobra.Command, args []string) error {
	if hrNav != review.NavHash && hrNav != review.NavReload {
		return fmt.Errorf("--nav must be %q or %q", review.NavHash, review.NavReload)
	}
	decks, err := reviewDecks(args)
	if err != nil {
		return err
	}
	switch {
	case hrSummary:
		return printReviewSummary(cmd, decks)
	case hrExport != "":
		return exportReview(decks[0])
	}
	return serveReview(cmd, decks)
}

func printReviewSummary(cmd *cobra.Command, decks []string) error {
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "%-44s %-8s %9s %8s %6s %11s %7s %9s\n", "Deck", "Language", "Approved", "Changes", "Stale", "Unreviewed", "Errors", "Warnings")
	for _, d := range decks {
		m, _, err := buildReviewModel(d, "", hrLanguages, true)
		if err != nil {
			return err
		}
		for i, l := range m.Languages {
			name := ""
			if i == 0 {
				name = truncate(m.Title, 44)
			}
			c := m.Summary[l]
			fmt.Fprintf(w, "%-44s %-8s %4d/%-4d %8d %6d %11d %7d %9d\n", name, l, c.Approved, c.Slides, c.Changes, c.Stale, c.Unreviewed, c.Errors, c.Warnings)
		}
	}
	return nil
}

func exportReview(deck string) error {
	deckAbs, err := filepath.Abs(deck)
	if err != nil {
		return err
	}
	outAbs, err := filepath.Abs(hrExport)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(filepath.Dir(outAbs), deckAbs)
	if err != nil {
		return fmt.Errorf("deck path relative to %s: %w", hrExport, err)
	}
	m, _, err := buildReviewModel(deck, urlPath(rel), hrLanguages, true)
	if err != nil {
		return err
	}
	f, err := os.Create(hrExport)
	if err != nil {
		return err
	}
	if err := review.Export(f, m); err != nil {
		_ = f.Close() // the export error is the one to report
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d slides, %v)\n", hrExport, len(m.Slides), m.Languages)
	return nil
}

// urlPath turns a relative file path into a relative URL path.
func urlPath(rel string) string {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

// deckFrameURL is where the review page loads a deck, relative to the page.
func deckFrameURL(deck string) string {
	u := strings.TrimPrefix(review.DeckPrefix, "/")
	if base := filepath.Base(deck); base != "index.html" {
		// The file server redirects index.html to its directory, which
		// would turn every fragment move into a reload.
		u += url.PathEscape(base)
	}
	return u
}

var nonSlugRe = regexp.MustCompile(`[^a-z0-9._-]+`)

// deckSlug names a deck in URLs: its file name, or for index.html its
// directory's name, made unique within used.
func deckSlug(deck string, used map[string]bool) string {
	name := strings.TrimSuffix(filepath.Base(deck), filepath.Ext(deck))
	if name == "index" {
		name = filepath.Base(filepath.Dir(deck))
	}
	slug := strings.Trim(nonSlugRe.ReplaceAllString(strings.ToLower(name), "-"), "-._")
	if slug == "" {
		slug = "deck"
	}
	for i, s := 2, slug; ; i++ {
		if !used[s] {
			used[s] = true
			return s
		}
		s = fmt.Sprintf("%s-%d", slug, i)
	}
}

// reviewHandlerOptions builds the handler options for one deck, creating its
// review file's directory.
func reviewHandlerOptions(deck string) (review.HandlerOptions, *review.Model, string, error) {
	frameURL := deckFrameURL(deck)
	m, in, err := buildReviewModel(deck, frameURL, hrLanguages, false)
	if err != nil {
		return review.HandlerOptions{}, nil, "", err
	}
	reviewPath := reviewFilePath(in.transcriptPath)
	if err := os.MkdirAll(filepath.Dir(reviewPath), 0o755); err != nil {
		return review.HandlerOptions{}, nil, "", err
	}
	return review.HandlerOptions{
		Load: func() (*review.Model, error) {
			m, _, err := buildReviewModel(deck, frameURL, hrLanguages, false)
			return m, err
		},
		ReviewPath: reviewPath,
		DeckDir:    filepath.Dir(deck),
		Reviewer:   hrReviewer,
	}, m, reviewPath, nil
}

func serveReview(cmd *cobra.Command, decks []string) error {
	w := cmd.OutOrStdout()
	var h http.Handler
	var multi []review.DeckOptions
	used := map[string]bool{}
	for _, d := range decks {
		opts, m, reviewPath, err := reviewHandlerOptions(d)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "%s · %d slides\n", m.Title, len(m.Slides))
		for _, l := range m.Languages {
			c := m.Summary[l]
			fmt.Fprintf(w, "  %s: %d of %d approved, %d errors, %d warnings\n", l, c.Approved, c.Slides, c.Errors, c.Warnings)
		}
		fmt.Fprintf(w, "  decisions: %s\n", reviewPath)
		if len(decks) == 1 {
			h = review.NewHandler(opts)
			break
		}
		multi = append(multi, review.DeckOptions{Slug: deckSlug(d, used), Name: m.Title, HandlerOptions: opts})
	}
	if len(multi) > 0 {
		var err error
		if h, err = review.NewMultiHandler(multi); err != nil {
			return err
		}
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", hrPort))
	if err != nil {
		return fmt.Errorf("listen on port %d: %w (try --port 0)", hrPort, err)
	}
	srv := &http.Server{Handler: h, ReadHeaderTimeout: 10 * time.Second}
	pageURL := "http://" + ln.Addr().String() + "/"
	fmt.Fprintf(w, "Review at %s (Ctrl+C to stop)\n", pageURL)
	if hrOpen {
		if err := openBrowser(pageURL); err != nil {
			fmt.Fprintf(os.Stderr, "could not open a browser: %v\n", err)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()
	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func openBrowser(u string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", u)
	case "windows":
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default:
		c = exec.Command("xdg-open", u)
	}
	return c.Start()
}

// requireReviewed returns an error unless every slide in slides (0-based
// indexes; nil for all) is approved on its current text in every language
// in langs.
func requireReviewed(slides []int, langs []string) error {
	m, in, err := buildReviewModel(hdInput, "", langs, true)
	if err != nil {
		return err
	}
	var problems []string
	for lang, nums := range m.Unready(langs) {
		var hit []string
		for _, n := range nums {
			if slides == nil || slices.Contains(slides, n-1) {
				hit = append(hit, fmt.Sprint(n))
			}
		}
		if len(hit) > 0 {
			problems = append(problems, fmt.Sprintf("%s: slides %s", lang, strings.Join(hit, ", ")))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	slices.Sort(problems)
	return fmt.Errorf("narration not approved in %s:\n  %s\nreview it with: vac html review --input %s",
		reviewFilePath(in.transcriptPath), strings.Join(problems, "\n  "), hdInput)
}
