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
	"github.com/grokify/videoascode/pkg/media"
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

A slide with a data-lesson attribute starts a lesson. Each lesson can be
recorded as its own video, for a course platform where one deck is a section
and each lesson is one video in it.

Subcommands:
  init     Write a transcript and video config for a deck
  lessons  List a deck's lessons and whether their videos are up to date
  video    Serve a deck locally and record it, or its lessons, with TTS voiceover

Examples:
  # Generate voiceover/transcript.json and deck.video.json next to the deck
  vac html init --input deck/index.html

  # Record the deck directly; no separate web server needed
  vac html video --input deck/index.html --output deck.mp4 --provider deepgram

  # Record one video per lesson into lessons/, with a lessons.json manifest
  vac html video --input deck/index.html --lessons all

  # Later: re-record only the lessons whose slides or narration changed
  vac html video --input deck/index.html --lessons changed`,
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
	hdLessonAttr     string
	hdLessonClass    string
	hdStartFragment  string
	hdOutputDir      string
)

// Flags for 'vac html video' only.
var hvLessons []string

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
	Short: "Serve an HTML deck locally and record it, or its lessons, with TTS voiceover",
	Long: `Serve the deck's directory on a local port, then record it one slide per
step with TTS voiceover, timing each slide to its own narration.

Narration comes from the deck itself, plus translations from the transcript
if one exists (see 'vac html init'). Accepts the same recording, TTS, and
subtitle flags as 'vac browser video'.

With --lessons, each selected lesson is recorded as its own video in
--output-dir (named like 02-designing-endpoints.mp4), and lessons.json there lists
every lesson with its videos, duration, and a fingerprint of the content it
was recorded from. --lessons takes lesson IDs or numbers, "all", or
"changed" (lessons not yet recorded or whose slides, narration, voice, or
pronunciations changed since). The deck is opened at each lesson's first
slide with --start-fragment (default "#{id}").

Voiceover clips are cached by their text and voice, so unchanged narration is
never sent to the TTS provider twice.`,
	Example: `  vac html video --input deck/index.html --output deck.mp4
  vac html video --input deck/index.html --output deck.mp4 \
    --provider deepgram --subtitles --lang en-US,fr-FR
  vac html video --input deck/index.html --lessons all --output-dir video/
  vac html video --input deck/index.html --lessons sec-design,3
  vac html video --input deck/index.html --lessons changed`,
	RunE: runHTMLVideo,
}

var htmlLessonsCmd = &cobra.Command{
	Use:   "lessons",
	Short: "List a deck's lessons and whether their videos are up to date",
	Long: `List the deck's lessons with their slides, estimated length, and status
against the lessons.json manifest in --output-dir:

  not recorded   no video yet
  changed        slides, narration, voice, or pronunciations changed since recording
  recorded       the video matches the deck`,
	Example: `  vac html lessons --input deck/index.html
  vac html lessons --input deck/index.html --output-dir video/`,
	RunE: runHTMLLessons,
}

func init() {
	for _, c := range []*cobra.Command{htmlInitCmd, htmlLessonsCmd, htmlVideoCmd} {
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
		c.Flags().StringVar(&hdLessonAttr, "lesson-attr", htmldeck.DefaultLessonAttr, "Attribute that marks a slide starting a lesson")
		c.Flags().StringVar(&hdLessonClass, "lesson-class", "", `Class that also marks a slide starting a lesson (e.g. "section")`)
		c.Flags().StringVar(&hdStartFragment, "start-fragment", htmldeck.DefaultStartFragment, "URL fragment that opens the deck at a lesson's first slide; {id}, {n} (1-based), {i} (0-based)")
		if err := c.MarkFlagRequired("input"); err != nil {
			panic(err)
		}
	}

	htmlInitCmd.Flags().StringVarP(&hiConfigPath, "config", "c", "", "Video config file to write (default: deck.video.json next to the deck)")
	htmlInitCmd.Flags().StringVar(&hiURL, "url", "", `URL the config records (default: "http://localhost:8000/<deck file>" plus --query)`)
	htmlInitCmd.Flags().StringVarP(&bvProvider, "provider", "p", "", "TTS provider recorded in the transcript: elevenlabs, deepgram, ... (default: keep the transcript's)")
	htmlInitCmd.Flags().StringVarP(&bvVoiceID, "voice", "v", "", "TTS voice ID recorded in the transcript (default: keep the transcript's)")

	addBrowserVideoFlags(htmlVideoCmd)
	htmlVideoCmd.Flags().StringSliceVar(&hvLessons, "lessons", nil, `Record lessons as separate videos: lesson IDs or numbers, "all", or "changed"`)
	for _, c := range []*cobra.Command{htmlLessonsCmd, htmlVideoCmd} {
		c.Flags().StringVar(&hdOutputDir, "output-dir", "", "Directory for lesson videos and lessons.json (default: lessons/ next to the deck)")
	}

	htmlCmd.AddCommand(htmlInitCmd, htmlLessonsCmd, htmlVideoCmd)
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
	deck, err := htmldeck.Parse(f, htmldeck.Options{
		SlideClass:    hdSlideClass,
		NarrationAttr: hdNarrationAttr,
		LessonAttr:    hdLessonAttr,
		LessonClass:   hdLessonClass,
	})
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

	pron, err := loadPronunciations(htmlConfigPath())
	if err != nil {
		return nil, err
	}
	return &htmlDeckInputs{deck: deck, transcript: t, transcriptPath: tPath, pronunciations: pron}, nil
}

// htmlConfigPath is the video config 'vac html init' writes.
func htmlConfigPath() string {
	if hiConfigPath != "" {
		return hiConfigPath
	}
	return filepath.Join(filepath.Dir(hdInput), "deck.video.json")
}

// loadPronunciations reads --pronunciations, or else keeps the
// pronunciations of the config at cfgPath, if it exists.
func loadPronunciations(cfgPath string) (map[string]map[string]string, error) {
	if hdPronunciations != "" {
		b, err := os.ReadFile(hdPronunciations)
		if err != nil {
			return nil, err
		}
		var pron map[string]map[string]string
		if err := json.Unmarshal(b, &pron); err != nil {
			return nil, fmt.Errorf("%s: %w", hdPronunciations, err)
		}
		return pron, nil
	}
	if _, err := os.Stat(cfgPath); errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	prev, err := config.LoadFromFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("read existing %s: %w", cfgPath, err)
	}
	return prev.Pronunciations, nil
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
		StartFragment:  hdStartFragment,
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

	cfgPath := htmlConfigPath()
	pron := in.pronunciations

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
	fmt.Printf("%d slides · %d lessons · ~%d words · ~%.1f min at 140 wpm · languages %v\n",
		len(in.deck.Slides), len(in.deck.Lessons()), words, float64(words)/140, cfg.Languages)
	if missing := in.deck.Unnarrated(); len(missing) > 0 {
		fmt.Printf("warning: slides without narration in %s: %v\n", hdLanguage, missing)
	}
	fmt.Printf("wrote %s\nwrote %s\n", in.transcriptPath, cfgPath)
	return nil
}

func runHTMLVideo(cmd *cobra.Command, args []string) error {
	ctx := newContext()
	if len(hvLessons) > 0 && cmd.Flags().Changed("output") {
		return fmt.Errorf("--output names a single video; with --lessons, use --output-dir")
	}
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
	if len(hvLessons) > 0 {
		return recordLessons(ctx, in, deckURL)
	}
	cfg, err := in.deck.VideoConfig(in.transcript, deckConfigOptions(deckURL, in.pronunciations))
	if err != nil {
		return err
	}
	return recordBrowserVideo(ctx, cfg, hdInput, bvOutputFile)
}

// lessonsDir is --output-dir, defaulting to lessons/ next to the deck.
func lessonsDir() string {
	if hdOutputDir != "" {
		return hdOutputDir
	}
	return filepath.Join(filepath.Dir(hdInput), "lessons")
}

// lessonPlan is a lesson with its video config, fingerprint, and status
// against the previous manifest.
type lessonPlan struct {
	lesson      htmldeck.Lesson
	cfg         *config.VideoConfig
	fingerprint string
	status      string
	entry       htmldeck.ManifestLesson
}

const (
	statusNotRecorded = "not recorded"
	statusChanged     = "changed"
	statusRecorded    = "recorded"
)

// planLessons builds each lesson's config and compares it with the manifest
// in --output-dir. Each plan's entry starts from the previous manifest's.
func planLessons(in *htmlDeckInputs, deckURL string) (*htmldeck.Manifest, []lessonPlan, error) {
	prev := &htmldeck.Manifest{}
	manifestPath := filepath.Join(lessonsDir(), htmldeck.ManifestFile)
	if _, err := os.Stat(manifestPath); err == nil {
		if prev, err = htmldeck.LoadManifest(manifestPath); err != nil {
			return nil, nil, err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, nil, err
	}

	ids := in.deck.SlideIDs()
	var plans []lessonPlan
	for _, l := range in.deck.Lessons() {
		cfg, err := in.deck.LessonVideoConfig(in.transcript, l, deckConfigOptions(deckURL, in.pronunciations))
		if err != nil {
			return nil, nil, fmt.Errorf("lesson %d (%s): %w", l.Number, l.ID, err)
		}
		p := lessonPlan{lesson: l, cfg: cfg, fingerprint: htmldeck.LessonFingerprint(l, cfg), status: statusNotRecorded}
		if old := prev.Lesson(l.ID); old != nil {
			p.entry = *old
			if videosExist(old.Videos) {
				p.status = statusRecorded
				if old.Fingerprint != p.fingerprint {
					p.status = statusChanged
				}
			}
		}
		p.entry.Number, p.entry.ID, p.entry.Title = l.Number, l.ID, l.Title
		p.entry.Slides = make([]string, len(l.Slides))
		for i, s := range l.Slides {
			p.entry.Slides[i] = ids[s.Index]
		}
		plans = append(plans, p)
	}
	return prev, plans, nil
}

func videosExist(videos map[string]string) bool {
	if len(videos) == 0 {
		return false
	}
	for _, v := range videos {
		if _, err := os.Stat(filepath.Join(lessonsDir(), v)); err != nil {
			return false
		}
	}
	return true
}

// selectLessons returns the indexes of the plans named by refs: lesson IDs
// or numbers, "all", or "changed".
func selectLessons(deck *htmldeck.Deck, plans []lessonPlan, refs []string) ([]int, error) {
	picked := make([]bool, len(plans))
	for _, ref := range refs {
		switch ref = strings.TrimSpace(ref); ref {
		case "all":
			for i := range picked {
				picked[i] = true
			}
		case "changed":
			for i, p := range plans {
				if p.status != statusRecorded {
					picked[i] = true
				}
			}
		default:
			l, err := deck.Lesson(ref)
			if err != nil {
				return nil, err
			}
			picked[l.Number-1] = true
		}
	}
	var out []int
	for i, ok := range picked {
		if ok {
			out = append(out, i)
		}
	}
	return out, nil
}

// recordLessons records the lessons selected by --lessons, one video each,
// and writes the manifest after each one so an interrupted run keeps the
// lessons already recorded.
func recordLessons(ctx context.Context, in *htmlDeckInputs, deckURL string) error {
	prev, plans, err := planLessons(in, deckURL)
	if err != nil {
		return err
	}
	selected, err := selectLessons(in.deck, plans, hvLessons)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		fmt.Println("All lessons are up to date.")
		return nil
	}
	if err := os.MkdirAll(lessonsDir(), 0o755); err != nil {
		return err
	}

	m := &htmldeck.Manifest{Title: in.transcript.Metadata.Title, DefaultLanguage: bvLanguages[0]}
	if m.Title == "" {
		m.Title = prev.Title
	}
	for _, p := range plans {
		m.Lessons = append(m.Lessons, p.entry)
	}
	manifestPath := filepath.Join(lessonsDir(), htmldeck.ManifestFile)

	for n, i := range selected {
		p := plans[i]
		fmt.Printf("\n[%d/%d] Lesson %d: %s (%d slides)\n", n+1, len(selected), p.lesson.Number, p.lesson.Title, len(p.lesson.Slides))
		file := p.lesson.Name() + ".mp4"
		out := filepath.Join(lessonsDir(), file)
		label := fmt.Sprintf("%s, lesson %d", hdInput, p.lesson.Number)
		if err := recordBrowserVideo(ctx, p.cfg, label, out); err != nil {
			return fmt.Errorf("lesson %d (%s): %w", p.lesson.Number, p.lesson.ID, err)
		}

		e := &m.Lessons[i]
		e.Videos = map[string]string{bvLanguages[0]: file}
		for _, lang := range bvLanguages[1:] {
			e.Videos[lang] = filepath.Base(languageVideoPath(out, lang))
		}
		if e.DurationMs, err = media.GetVideoDurationMs(out); err != nil {
			return fmt.Errorf("lesson %d: read duration: %w", p.lesson.Number, err)
		}
		e.Fingerprint = p.fingerprint
		if err := m.Save(manifestPath); err != nil {
			return err
		}
	}
	fmt.Printf("\nRecorded %d of %d lessons; manifest: %s\n", len(selected), len(plans), manifestPath)
	return nil
}

func runHTMLLessons(cmd *cobra.Command, args []string) error {
	in, err := loadHTMLDeck()
	if err != nil {
		return err
	}
	// The URL is not part of a lesson's fingerprint, so any URL will do.
	_, plans, err := planLessons(in, "http://localhost/"+url.PathEscape(filepath.Base(hdInput)))
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "%s · %d slides · %d lessons\n\n", in.transcript.Metadata.Title, len(in.deck.Slides), len(plans))
	fmt.Fprintf(w, "%3s  %-22s %-34s %6s %6s  %s\n", "#", "ID", "Title", "Slides", "~Min", "Status")
	for _, p := range plans {
		words, pause := 0, 0
		for _, s := range p.lesson.Slides {
			words += len(strings.Fields(s.Narration))
			pause += s.Pause
		}
		minutes := float64(words)/140 + float64(pause)/60000
		status := p.status
		if p.status != statusNotRecorded && p.entry.DurationMs > 0 {
			d := time.Duration(p.entry.DurationMs) * time.Millisecond
			status += fmt.Sprintf(" (%d:%02d)", int(d.Minutes()), int(d.Seconds())%60)
		}
		fmt.Fprintf(w, "%3d  %-22s %-34s %6d %6.1f  %s\n", p.lesson.Number, truncate(p.lesson.ID, 22), truncate(p.lesson.Title, 34), len(p.lesson.Slides), minutes, status)
	}
	return nil
}

func truncate(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n-1]) + "…"
	}
	return s
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
