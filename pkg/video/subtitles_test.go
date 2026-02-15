package video

import "testing"

func TestBCP47ToISO639(t *testing.T) {
	tests := []struct {
		name     string
		bcp47    string
		expected string
	}{
		// English variants
		{"English base", "en", "eng"},
		{"English US", "en-US", "eng"},
		{"English GB", "en-GB", "eng"},
		{"English AU", "en-AU", "eng"},

		// French variants
		{"French base", "fr", "fra"},
		{"French FR", "fr-FR", "fra"},
		{"French CA", "fr-CA", "fra"},

		// German variants
		{"German base", "de", "deu"},
		{"German DE", "de-DE", "deu"},

		// Spanish variants
		{"Spanish base", "es", "spa"},
		{"Spanish ES", "es-ES", "spa"},
		{"Spanish MX", "es-MX", "spa"},

		// Chinese variants
		{"Chinese base", "zh", "zho"},
		{"Chinese Simplified", "zh-Hans", "zho"},
		{"Chinese Traditional", "zh-Hant", "zho"},
		{"Chinese CN", "zh-CN", "zho"},
		{"Chinese TW", "zh-TW", "zho"},

		// Japanese
		{"Japanese base", "ja", "jpn"},
		{"Japanese JP", "ja-JP", "jpn"},

		// Korean
		{"Korean base", "ko", "kor"},
		{"Korean KR", "ko-KR", "kor"},

		// Portuguese variants
		{"Portuguese base", "pt", "por"},
		{"Portuguese BR", "pt-BR", "por"},
		{"Portuguese PT", "pt-PT", "por"},

		// Italian
		{"Italian base", "it", "ita"},
		{"Italian IT", "it-IT", "ita"},

		// Russian
		{"Russian base", "ru", "rus"},
		{"Russian RU", "ru-RU", "rus"},

		// Arabic
		{"Arabic", "ar", "ara"},

		// Hindi
		{"Hindi base", "hi", "hin"},
		{"Hindi IN", "hi-IN", "hin"},

		// Already ISO 639-2
		{"Already ISO eng", "eng", "eng"},
		{"Already ISO fra", "fra", "fra"},
		{"Already ISO zho", "zho", "zho"},

		// Unknown - should return truncated/lowercased
		{"Unknown with region", "xx-YY", "xx"},
		{"Unknown base", "xyz", "xyz"},

		// Edge cases
		{"Empty string", "", "und"},
		{"Single char", "e", "und"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BCP47ToISO639(tt.bcp47)
			if result != tt.expected {
				t.Errorf("BCP47ToISO639(%q) = %q, want %q", tt.bcp47, result, tt.expected)
			}
		})
	}
}

func TestDetectLanguageFromSubtitlePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		// Simple locale filename
		{"Simple en-US", "en-US.srt", "en-US"},
		{"Simple fr-FR", "fr-FR.vtt", "fr-FR"},
		{"Simple zh-Hans", "zh-Hans.srt", "zh-Hans"},

		// With directory
		{"With dir en-US", "subtitles/en-US.srt", "en-US"},
		{"With dir fr-FR", "subtitles/fr-FR.vtt", "fr-FR"},
		{"Full path", "/path/to/subtitles/zh-Hans.srt", "zh-Hans"},

		// With prefix
		{"With prefix", "slide_001.en-US.srt", "en-US"},
		{"Combined prefix", "combined.fr-FR.vtt", "fr-FR"},

		// No locale detected
		{"No locale simple", "subtitles.srt", ""},
		{"No locale numbered", "slide_001.srt", ""},
		{"No locale combined", "combined.srt", ""},

		// Edge cases
		{"Empty", "", ""},
		{"Just extension", ".srt", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectLanguageFromSubtitlePath(tt.path)
			if result != tt.expected {
				t.Errorf("DetectLanguageFromSubtitlePath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsValidLocale(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid locales
		{"en-US", "en-US", true},
		{"fr-FR", "fr-FR", true},
		{"de-DE", "de-DE", true},
		{"zh-Hans", "zh-Hans", true},
		{"zh-Hant", "zh-Hant", true},
		{"pt-BR", "pt-BR", true},

		// Invalid - no hyphen
		{"No hyphen", "enUS", false},
		{"Just language", "en", false},

		// Invalid - wrong format
		{"Lowercase region", "en-us", false},
		{"Uppercase language", "EN-US", false},
		{"Wrong script case", "zh-hans", false},
		{"Too short", "e-US", false},
		{"Too long language", "engl-US", false},
		{"Wrong region length", "en-USA", false},

		// Invalid - too many parts
		{"Three parts", "en-US-extra", false},

		// Edge cases
		{"Empty", "", false},
		{"Just hyphen", "-", false},
		{"Hyphen at start", "-US", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidLocale(tt.input)
			if result != tt.expected {
				t.Errorf("isValidLocale(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
