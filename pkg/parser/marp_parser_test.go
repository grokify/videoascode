package parser

import (
	"testing"
)

func TestParseMarpFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		wantSlides  int
		checkSlides func(*testing.T, *Presentation)
	}{
		{
			name: "basic presentation with two slides",
			content: `---
marp: true
theme: default
---

# Slide 1

<!-- This is the voiceover for slide 1 -->

Content here

---

# Slide 2

<!-- This is the voiceover for slide 2 -->

More content
`,
			wantErr:    false,
			wantSlides: 2,
			checkSlides: func(t *testing.T, p *Presentation) {
				if p.Slides[0].Voiceover != "This is the voiceover for slide 1" {
					t.Errorf("Slide 0 voiceover = %q, want %q", p.Slides[0].Voiceover, "This is the voiceover for slide 1")
				}
				if p.Slides[1].Voiceover != "This is the voiceover for slide 2" {
					t.Errorf("Slide 1 voiceover = %q, want %q", p.Slides[1].Voiceover, "This is the voiceover for slide 2")
				}
			},
		},
		{
			name: "slide with pause directive",
			content: `---
marp: true
---

# Slide 1

<!-- Hello world [PAUSE:1000] and goodbye -->
`,
			wantErr:    false,
			wantSlides: 1,
			checkSlides: func(t *testing.T, p *Presentation) {
				if p.Slides[0].Voiceover != "Hello world  and goodbye" {
					t.Errorf("Slide 0 voiceover = %q, want %q", p.Slides[0].Voiceover, "Hello world  and goodbye")
				}
				if p.Slides[0].TotalPauseDuration != 1000 {
					t.Errorf("Slide 0 TotalPauseDuration = %d, want 1000", p.Slides[0].TotalPauseDuration)
				}
				if len(p.Slides[0].PauseDurations) != 1 || p.Slides[0].PauseDurations[0] != 1000 {
					t.Errorf("Slide 0 PauseDurations = %v, want [1000]", p.Slides[0].PauseDurations)
				}
			},
		},
		{
			name: "slide with multiple pause directives",
			content: `---
marp: true
---

# Slide 1

<!-- First part [PAUSE:500] middle [PAUSE:1500] end -->
`,
			wantErr:    false,
			wantSlides: 1,
			checkSlides: func(t *testing.T, p *Presentation) {
				if p.Slides[0].TotalPauseDuration != 2000 {
					t.Errorf("Slide 0 TotalPauseDuration = %d, want 2000", p.Slides[0].TotalPauseDuration)
				}
				if len(p.Slides[0].PauseDurations) != 2 {
					t.Errorf("Slide 0 PauseDurations length = %d, want 2", len(p.Slides[0].PauseDurations))
				}
			},
		},
		{
			name: "slide with no voiceover",
			content: `---
marp: true
---

# Slide 1

Just content, no comments
`,
			wantErr:    false,
			wantSlides: 1,
			checkSlides: func(t *testing.T, p *Presentation) {
				if p.Slides[0].Voiceover != "" {
					t.Errorf("Slide 0 voiceover = %q, want empty string", p.Slides[0].Voiceover)
				}
			},
		},
		{
			name: "slide with multiple comments",
			content: `---
marp: true
---

# Slide 1

<!-- First comment -->
<!-- Second comment -->
`,
			wantErr:    false,
			wantSlides: 1,
			checkSlides: func(t *testing.T, p *Presentation) {
				want := "First comment\nSecond comment"
				if p.Slides[0].Voiceover != want {
					t.Errorf("Slide 0 voiceover = %q, want %q", p.Slides[0].Voiceover, want)
				}
			},
		},
		{
			name:       "missing frontmatter",
			content:    "# Just content\nNo frontmatter here",
			wantErr:    true,
			wantSlides: 0,
		},
		{
			name: "frontmatter is preserved",
			content: `---
marp: true
theme: gaia
paginate: true
---

# Content
`,
			wantErr:    false,
			wantSlides: 1,
			checkSlides: func(t *testing.T, p *Presentation) {
				if p.Frontmatter == "" {
					t.Error("Frontmatter should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMarpFile(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMarpFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got.Slides) != tt.wantSlides {
				t.Errorf("ParseMarpFile() got %d slides, want %d", len(got.Slides), tt.wantSlides)
			}
			if tt.checkSlides != nil {
				tt.checkSlides(t, got)
			}
		})
	}
}

func TestParseSlide(t *testing.T) {
	tests := []struct {
		name               string
		index              int
		text               string
		wantVoiceover      string
		wantPauseDurations []int
		wantTotalPause     int
	}{
		{
			name:               "simple slide with voiceover",
			index:              0,
			text:               "# Title\n\n<!-- This is voiceover -->\n\nContent",
			wantVoiceover:      "This is voiceover",
			wantPauseDurations: []int{},
			wantTotalPause:     0,
		},
		{
			name:               "slide with pause",
			index:              1,
			text:               "# Title\n\n<!-- Hello [PAUSE:2000] world -->",
			wantVoiceover:      "Hello  world",
			wantPauseDurations: []int{2000},
			wantTotalPause:     2000,
		},
		{
			name:               "empty slide",
			index:              0,
			text:               "",
			wantVoiceover:      "",
			wantPauseDurations: []int{},
			wantTotalPause:     0,
		},
		{
			name:               "invalid pause directive ignored",
			index:              0,
			text:               "<!-- Hello [PAUSE:notanumber] world -->",
			wantVoiceover:      "Hello [PAUSE:notanumber] world",
			wantPauseDurations: []int{},
			wantTotalPause:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSlide(tt.index, tt.text)

			if got.Index != tt.index {
				t.Errorf("parseSlide() Index = %d, want %d", got.Index, tt.index)
			}
			if got.Voiceover != tt.wantVoiceover {
				t.Errorf("parseSlide() Voiceover = %q, want %q", got.Voiceover, tt.wantVoiceover)
			}
			if got.TotalPauseDuration != tt.wantTotalPause {
				t.Errorf("parseSlide() TotalPauseDuration = %d, want %d", got.TotalPauseDuration, tt.wantTotalPause)
			}
			if len(got.PauseDurations) != len(tt.wantPauseDurations) {
				t.Errorf("parseSlide() PauseDurations = %v, want %v", got.PauseDurations, tt.wantPauseDurations)
			}
		})
	}
}
