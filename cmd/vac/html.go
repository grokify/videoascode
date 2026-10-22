package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/grokify/videoascode/pkg/config"
	"github.com/grokify/videoascode/pkg/htmldeck"
	"github.com/grokify/videoascode/pkg/transcript"
	"github.com/spf13/cobra"
)

var htmlCmd = &cobra.Command{
	Use:   "html",
	Short: "Commands for native HTML slide decks",
	Long: `Commands for narrating and recording HTML slide decks.

A deck is an HTML page whose slides carry a "slide" class and advance with a
key press (ArrowRight by default). Each slide's narration comes from its
data-vo attribute, or from a child <aside class="notes"> element (reveal.js
speaker notes). [PAUSE:ms] directives add silence after a slide.

Subcommands:
  init     Write a transcript and video config for a deck
  video    Serve a deck locally and record it with TTS voiceover

Examples:
  # Generate voiceover/transcript.json and deck.video.json next to the deck
  vac html init --input deck/index.html

  # Record the deck directly; no separate web server needed
  vac html video --input deck/index.html --output deck.mp4 --provider deepgram`,
}

// Flags shared by 'vac html init' and 'vac html video'.
var (
	hdInput          string
	hdTitle          string
	hdLanguage       string
	hdTranscriptPath string
	hdPronunciations string
	hdNextKey        string
	hdQuery          string
	hdInitialWait    int
	hdSlideClass     string
	hdNarrationAttr  string
	hdVoiceName      string
	hdVoiceModel     string
)

// Flags for 'vac html init' only.
var (
	hiConfigPath string
	hiURL        string
)

var htmlInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Write a transcript and video config for an HTML deck",
	Long: `Read an HTML deck's slide narration and write:

  - a transcript (videoascode transcript format), keeping any translations
    already in it, so it can be edited and translated over time
  - a browser video config for 'vac browser video', with one step per slide
    and narration for every language that covers all slides

Re-run it after editing narration; translations and pronunciations in the
existing files are kept.`,
	Example: `  vac html init --input deck/index.html
  vac html init --input deck/index.html --title "My Talk" \
    --provider deepgram --voice aura-asteria-en \
    --url "http://localhost:8000/index.html?capture=1"`,
	RunE: runHTMLInit,
}

var htmlVideoCmd = &cobra.Command{
	Use:   "video",
	Short: "Serve an HTML deck locally and record it with TTS voiceover",
	Long: `Serve the deck's directory on a local port, then record it one slide per
step with TTS voiceover, timing each slide to its own narration.

Narration comes from the deck itself, plus translations from the transcript
if one exists (see 'vac html init'). Accepts the same recording, TTS, and
subtitle flags as 'vac browser video'.`,
	Example: `  vac html video --input deck/index.html --output deck.mp4
  vac html video --input deck/index.html --output deck.mp4 \
    --provider deepgram --subtitles --lang en-US,fr-FR`,
	RunE: runHTMLVideo,
}

func init() {
	for _, c := range []*cobra.Command{htmlInitCmd, htmlVideoCmd} {
		c.Flags().StringVarP(&hdInput, "input", "i", "", "HTML deck file (required)")
		c.Flags().StringVar(&hdTitle, "title", "", "Title (default: the transcript's, else the page <title>)")
		c.Flags().StringVar(&hdLanguage, "narration-lang", htmldeck.DefaultLanguage, "Language of the narration in the HTML")
		c.Flags().StringVar(&hdTranscriptPath, "transcript", "", "Transcript file (default: voiceover/transcript.json next to the deck)")
		c.Flags().StringVar(&hdPronunciations, "pronunciations", "", `JSON file mapping terms to spoken forms per language, e.g. {"vac": {"en-US": "V A C"}}`)
		c.Flags().StringVar(&hdNextKey, "next-key", htmldeck.DefaultNextKey, "Key that advances one slide")
		c.Flags().StringVar(&hdQuery, "query", "capture=1", "Query string added to the deck URL (e.g. to hide on-screen controls); empty for none")
		c.Flags().IntVar(&hdInitialWait, "initial-wait", htmldeck.DefaultInitialWaitMs, "Milliseconds to wait on the first slide before recording")
		c.Flags().StringVar(&hdSlideClass, "slide-class", htmldeck.DefaultSlideClass, "Class that marks a slide element")
		c.Flags().StringVar(&hdNarrationAttr, "narration-attr", htmldeck.DefaultNarrationAttr, "Attribute holding a slide's narration")
		c.Flags().StringVar(&hdVoiceName, "voice-name", "", "Human-readable voice name for the transcript")
		c.Flags().StringVar(&hdVoiceModel, "voice-model", "", "Provider-specific TTS model for the transcript")
		if err := c.MarkFlagRequired("input"); err != nil {
			panic(err)
		}
	}

	htmlInitCmd.Flags().StringVarP(&hiConfigPath, "config", "c", "", "Video config file to write (default: deck.video.json next to the deck)")
	htmlInitCmd.Flags().StringVar(&hiURL, "url", "", `URL the config records (default: "http://localhost:8000/<deck file>" plus --query)`)
	htmlInitCmd.Flags().StringVarP(&bvProvider, "provider", "p", "", "TTS provider recorded in the transcript: elevenlabs, deepgram, ... (default: keep the transcript's)")
	htmlInitCmd.Flags().StringVarP(&bvVoiceID, "voice", "v", "", "TTS voice ID recorded in the transcript (default: keep the transcript's)")

	addBrowserVideoFlags(htmlVideoCmd)

	htmlCmd.AddCommand(htmlInitCmd, htmlVideoCmd)
	rootCmd.AddCommand(htmlCmd)
}

// htmlDeckInputs is a parsed deck with its transcript and pronunciations.
type htmlDeckInputs struct {
	deck           *htmldeck.Deck
	transcript     *transcript.Transcript
	transcriptPath string
	pronunciations map[string]map[string]string
}

// loadHTMLDeck parses the deck and builds its transcript, keeping the
// translations and voice of any existing transcript.
func loadHTMLDeck() (*htmlDeckInputs, error) {
	f, err := os.Open(hdInput)
	if err != nil {
		return nil, err
	}
	deck, err := htmldeck.Parse(f, htmldeck.Options{SlideClass: hdSlideClass, NarrationAttr: hdNarrationAttr})
	closeErr := f.Close()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", hdInput, err)
	}
	if closeErr != nil {
		return nil, closeErr
	}

	tPath := hdTranscriptPath
	if tPath == "" {
		tPath = filepath.Join(filepath.Dir(hdInput), "voiceover", "transcript.json")
	}
	existing, err := loadTranscriptIfExists(tPath)
	if err != nil {
		return nil, err
	}

	voice := transcript.VoiceConfig{Provider: bvProvider, VoiceID: bvVoiceID, VoiceName: hdVoiceName, Model: hdVoiceModel}
	if existing != nil && bvProvider == "" && bvVoiceID == "" {
		voice = existing.Metadata.DefaultVoice
	}
	t := deck.Transcript(htmldeck.TranscriptOptions{
		Title:    hdTitle,
		Language: hdLanguage,
		Voice:    voice,
		Existing: existing,
	})

	var pron map[string]map[string]string
	if hdPronunciations != "" {
		b, err := os.ReadFile(hdPronunciations)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &pron); err != nil {
			return nil, fmt.Errorf("%s: %w", hdPronunciations, err)
		}
	}
	return &htmlDeckInputs{deck: deck, transcript: t, transcriptPath: tPath, pronunciations: pron}, nil
}

func loadTranscriptIfExists(path string) (*transcript.Transcript, error) {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return transcript.LoadFromFile(path)
}

func withQuery(u string) string {
	if hdQuery == "" {
		return u
	}
	sep := "?"
	if strings.Contains(u, "?") {
		sep = "&"
	}
	return u + sep + hdQuery
}

func deckConfigOptions(deckURL string, pron map[string]map[string]string) htmldeck.ConfigOptions {
	return htmldeck.ConfigOptions{
		URL:            deckURL,
		NextKey:        hdNextKey,
		InitialWaitMs:  hdInitialWait,
		Width:          bvWidth,
		Height:         bvHeight,
		FrameRate:      bvFPS,
		Headless:       bvHeadless,
		Pronunciations: pron,
	}
}

func runHTMLInit(cmd *cobra.Command, args []string) error {
	in, err := loadHTMLDeck()
	if err != nil {
		return err
	}

	cfgPath := hiConfigPath
	if cfgPath == "" {
		cfgPath = filepath.Join(filepath.Dir(hdInput), "deck.video.json")
	}
	pron := in.pronunciations
	if pron == nil {
		// Keep pronunciations from a config written earlier.
		if _, statErr := os.Stat(cfgPath); statErr == nil {
			prev, err := config.LoadFromFile(cfgPath)
			if err != nil {
				return fmt.Errorf("read existing %s: %w", cfgPath, err)
			}
			pron = prev.Pronunciations
		}
	}

	deckURL := hiURL
	if deckURL == "" {
		deckURL = withQuery("http://localhost:8000/" + url.PathEscape(filepath.Base(hdInput)))
	}
	cfg, err := in.deck.VideoConfig(in.transcript, deckConfigOptions(deckURL, pron))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(in.transcriptPath), 0o755); err != nil {
		return err
	}
	if err := in.transcript.SaveToFile(in.transcriptPath); err != nil {
		return err
	}
	if err := cfg.SaveToFile(cfgPath); err != nil {
		return err
	}

	words := 0
	for _, s := range in.deck.Slides {
		words += len(strings.Fields(s.Narration))
	}
	fmt.Printf("%d slides · ~%d words · ~%.1f min at 140 wpm · languages %v\n",
		len(in.deck.Slides), words, float64(words)/140, cfg.Languages)
	if missing := in.deck.Unnarrated(); len(missing) > 0 {
		fmt.Printf("warning: slides without narration in %s: %v\n", hdLanguage, missing)
	}
	fmt.Printf("wrote %s\nwrote %s\n", in.transcriptPath, cfgPath)
	return nil
}

func runHTMLVideo(cmd *cobra.Command, args []string) error {
	ctx := newContext()
	in, err := loadHTMLDeck()
	if err != nil {
		return err
	}

	baseURL, stop, err := serveDir(filepath.Dir(hdInput))
	if err != nil {
		return err
	}
	defer stop()

	deckURL := withQuery(baseURL + "/" + url.PathEscape(filepath.Base(hdInput)))
	cfg, err := in.deck.VideoConfig(in.transcript, deckConfigOptions(deckURL, in.pronunciations))
	if err != nil {
		return err
	}
	return recordBrowserVideo(ctx, cfg, hdInput)
}

// serveDir serves dir on a free loopback port and returns its base URL and a
// function that shuts the server down.
func serveDir(dir string) (string, func(), error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, fmt.Errorf("listen: %w", err)
	}
	srv := &http.Server{
		Handler:           http.FileServerFS(os.DirFS(dir)),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "deck server: %v\n", err)
		}
	}()
	stop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "deck server shutdown: %v\n", err)
		}
	}
	return "http://" + ln.Addr().String(), stop, nil
}
