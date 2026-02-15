package tts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Dictionary contains case corrections for subtitle text.
type Dictionary struct {
	Name        string            `json:"name,omitempty"`
	Version     string            `json:"version,omitempty"`
	Corrections map[string]string `json:"corrections"`
}

// DictionaryLoader handles loading and merging multiple dictionaries.
type DictionaryLoader struct {
	configDir        string
	projectDir       string
	additionalPaths  []string
	includeBuiltIn   bool
}

// NewDictionaryLoader creates a new dictionary loader.
func NewDictionaryLoader() *DictionaryLoader {
	// Default config directory
	configDir := ""
	if home, err := os.UserHomeDir(); err == nil {
		configDir = filepath.Join(home, ".config", "marp2video", "dictionaries")
	}

	return &DictionaryLoader{
		configDir:      configDir,
		projectDir:     "./dictionaries",
		includeBuiltIn: true,
	}
}

// WithConfigDir sets a custom config directory.
func (dl *DictionaryLoader) WithConfigDir(dir string) *DictionaryLoader {
	dl.configDir = dir
	return dl
}

// WithProjectDir sets the project dictionary directory.
func (dl *DictionaryLoader) WithProjectDir(dir string) *DictionaryLoader {
	dl.projectDir = dir
	return dl
}

// WithAdditionalPaths adds extra dictionary paths.
func (dl *DictionaryLoader) WithAdditionalPaths(paths []string) *DictionaryLoader {
	dl.additionalPaths = paths
	return dl
}

// WithBuiltIn controls whether to include built-in corrections.
func (dl *DictionaryLoader) WithBuiltIn(include bool) *DictionaryLoader {
	dl.includeBuiltIn = include
	return dl
}

// Load loads and merges all dictionaries in order.
// Order: built-in → user config → additional paths → project local
func (dl *DictionaryLoader) Load() (*Dictionary, error) {
	merged := &Dictionary{
		Name:        "merged",
		Corrections: make(map[string]string),
	}

	// 1. Built-in corrections
	if dl.includeBuiltIn {
		for k, v := range builtInCorrections {
			merged.Corrections[k] = v
		}
	}

	// 2. User config directory (~/.config/marp2video/dictionaries/*.json)
	if dl.configDir != "" {
		if err := dl.loadFromDir(dl.configDir, merged); err != nil {
			// Don't fail if config dir doesn't exist
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load config dictionaries: %w", err)
			}
		}
	}

	// 3. Additional paths (--dictionary flags)
	for _, path := range dl.additionalPaths {
		if err := dl.loadFromPath(path, merged); err != nil {
			return nil, fmt.Errorf("failed to load dictionary %s: %w", path, err)
		}
	}

	// 4. Project local directory (./dictionaries/*.json)
	if dl.projectDir != "" {
		if err := dl.loadFromDir(dl.projectDir, merged); err != nil {
			// Don't fail if project dir doesn't exist
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load project dictionaries: %w", err)
			}
		}
	}

	return merged, nil
}

// loadFromDir loads all .json files from a directory.
func (dl *DictionaryLoader) loadFromDir(dir string, merged *Dictionary) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Sort entries to ensure consistent order
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if err := dl.loadFromPath(path, merged); err != nil {
			return err
		}
	}

	return nil
}

// loadFromPath loads a single dictionary file and merges it.
func (dl *DictionaryLoader) loadFromPath(path string, merged *Dictionary) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var dict Dictionary
	if err := json.Unmarshal(data, &dict); err != nil {
		return fmt.Errorf("invalid JSON in %s: %w", path, err)
	}

	// Merge corrections (later values override earlier)
	for k, v := range dict.Corrections {
		// Normalize key to lowercase for consistent lookup
		merged.Corrections[strings.ToLower(k)] = v
	}

	return nil
}

// CaseCorrector applies dictionary-based case corrections to text.
type CaseCorrector struct {
	dictionary *Dictionary
	patterns   []*correctionPattern
}

type correctionPattern struct {
	original    string
	replacement string
	regex       *regexp.Regexp
	wordCount   int
}

// NewCaseCorrector creates a new case corrector from a dictionary.
func NewCaseCorrector(dict *Dictionary) *CaseCorrector {
	cc := &CaseCorrector{
		dictionary: dict,
		patterns:   make([]*correctionPattern, 0, len(dict.Corrections)),
	}

	// Build patterns sorted by word count (longer phrases first)
	for original, replacement := range dict.Corrections {
		wordCount := len(strings.Fields(original))
		// Build regex that matches word boundaries (case-insensitive)
		pattern := `(?i)\b` + regexp.QuoteMeta(original) + `\b`
		regex, err := regexp.Compile(pattern)
		if err != nil {
			continue // Skip invalid patterns
		}
		cc.patterns = append(cc.patterns, &correctionPattern{
			original:    original,
			replacement: replacement,
			regex:       regex,
			wordCount:   wordCount,
		})
	}

	// Sort by word count descending (match longer phrases first)
	sort.Slice(cc.patterns, func(i, j int) bool {
		return cc.patterns[i].wordCount > cc.patterns[j].wordCount
	})

	return cc
}

// Correct applies all dictionary corrections to the text.
func (cc *CaseCorrector) Correct(text string) string {
	result := text
	for _, p := range cc.patterns {
		result = p.regex.ReplaceAllString(result, p.replacement)
	}
	return result
}

// CorrectWord corrects a single word using the dictionary.
func (cc *CaseCorrector) CorrectWord(word string) string {
	lower := strings.ToLower(word)

	// Check for exact match
	if replacement, ok := cc.dictionary.Corrections[lower]; ok {
		// Preserve trailing punctuation
		return preservePunctuation(word, replacement)
	}

	// Check without punctuation
	stripped := stripTrailingPunctuation(lower)
	if replacement, ok := cc.dictionary.Corrections[stripped]; ok {
		return preservePunctuation(word, replacement)
	}

	return word
}

// stripTrailingPunctuation removes trailing punctuation from a word.
func stripTrailingPunctuation(word string) string {
	return strings.TrimRight(word, ".,!?;:'\"")
}

// preservePunctuation applies the replacement but keeps original punctuation.
func preservePunctuation(original, replacement string) string {
	// Find trailing punctuation in original
	trailing := ""
	for i := len(original) - 1; i >= 0; i-- {
		c := original[i]
		if c == '.' || c == ',' || c == '!' || c == '?' || c == ';' || c == ':' || c == '\'' || c == '"' {
			trailing = string(c) + trailing
		} else {
			break
		}
	}
	return replacement + trailing
}

// builtInCorrections contains common technical terms with proper capitalization.
var builtInCorrections = map[string]string{
	// Pronouns
	"i": "I",

	// AI/ML terms
	"ai":               "AI",
	"ml":               "ML",
	"llm":              "LLM",
	"llms":             "LLMs",
	"gpt":              "GPT",
	"nlp":              "NLP",
	"rag":              "RAG",
	"genai":            "GenAI",
	"generative ai":    "Generative AI",
	"machine learning": "Machine Learning",
	"deep learning":    "Deep Learning",
	"neural network":   "Neural Network",
	"neural networks":  "Neural Networks",

	// AI Companies & Products
	"openai":       "OpenAI",
	"chatgpt":      "ChatGPT",
	"claude":       "Claude",
	"claude code":  "Claude Code",
	"anthropic":    "Anthropic",
	"gemini":       "Gemini",
	"copilot":      "Copilot",
	"github copilot": "GitHub Copilot",
	"midjourney":   "Midjourney",
	"dall-e":       "DALL-E",
	"dalle":        "DALL-E",
	"stable diffusion": "Stable Diffusion",
	"hugging face": "Hugging Face",
	"huggingface":  "HuggingFace",
	"langchain":    "LangChain",
	"llamaindex":   "LlamaIndex",
	"llama":        "LLaMA",
	"mistral":      "Mistral",
	"deepseek":     "DeepSeek",
	"perplexity":   "Perplexity",
	"cursor":       "Cursor",
	"windsurf":     "Windsurf",
	"codeium":      "Codeium",
	"tabnine":      "Tabnine",
	"replit":       "Replit",
	"v0":           "v0",
	"bolt":         "Bolt",

	// Tech Companies
	"google":    "Google",
	"microsoft": "Microsoft",
	"amazon":    "Amazon",
	"aws":       "AWS",
	"meta":      "Meta",
	"apple":     "Apple",
	"nvidia":    "NVIDIA",
	"intel":     "Intel",
	"amd":       "AMD",
	"ibm":       "IBM",
	"oracle":    "Oracle",
	"salesforce": "Salesforce",
	"slack":     "Slack",
	"zoom":      "Zoom",
	"notion":    "Notion",
	"figma":     "Figma",
	"vercel":    "Vercel",
	"netlify":   "Netlify",
	"cloudflare": "Cloudflare",
	"datadog":   "Datadog",
	"splunk":    "Splunk",
	"elastic":   "Elastic",
	"mongodb":   "MongoDB",
	"redis":     "Redis",
	"postgres":  "Postgres",
	"postgresql": "PostgreSQL",
	"mysql":     "MySQL",
	"sqlite":    "SQLite",

	// Developer Tools & Platforms
	"github":     "GitHub",
	"gitlab":     "GitLab",
	"bitbucket":  "Bitbucket",
	"git":        "Git",
	"docker":     "Docker",
	"kubernetes": "Kubernetes",
	"k8s":        "K8s",
	"terraform":  "Terraform",
	"jenkins":    "Jenkins",
	"circleci":   "CircleCI",
	"travis":     "Travis",
	"jira":       "Jira",
	"confluence": "Confluence",
	"vscode":     "VSCode",
	"vs code":    "VS Code",
	"visual studio code": "Visual Studio Code",
	"intellij":   "IntelliJ",
	"pycharm":    "PyCharm",
	"webstorm":   "WebStorm",
	"vim":        "Vim",
	"neovim":     "Neovim",
	"emacs":      "Emacs",

	// Programming Languages
	"javascript": "JavaScript",
	"typescript": "TypeScript",
	"python":     "Python",
	"golang":     "Golang",
	"rust":       "Rust",
	"java":       "Java",
	"kotlin":     "Kotlin",
	"swift":      "Swift",
	"c++":        "C++",
	"c#":         "C#",
	"ruby":       "Ruby",
	"php":        "PHP",
	"scala":      "Scala",
	"haskell":    "Haskell",
	"elixir":     "Elixir",
	"clojure":    "Clojure",

	// Web Technologies
	"html":       "HTML",
	"css":        "CSS",
	"json":       "JSON",
	"xml":        "XML",
	"yaml":       "YAML",
	"graphql":    "GraphQL",
	"rest":       "REST",
	"restful":    "RESTful",
	"api":        "API",
	"apis":       "APIs",
	"sdk":        "SDK",
	"sdks":       "SDKs",
	"cli":        "CLI",
	"gui":        "GUI",
	"ui":         "UI",
	"ux":         "UX",
	"http":       "HTTP",
	"https":      "HTTPS",
	"url":        "URL",
	"urls":       "URLs",
	"uri":        "URI",
	"oauth":      "OAuth",
	"jwt":        "JWT",
	"saml":       "SAML",
	"sso":        "SSO",
	"mfa":        "MFA",
	"totp":       "TOTP",

	// Frameworks & Libraries
	"react":      "React",
	"reactjs":    "ReactJS",
	"vue":        "Vue",
	"vuejs":      "VueJS",
	"angular":    "Angular",
	"svelte":     "Svelte",
	"nextjs":     "Next.js",
	"next.js":    "Next.js",
	"nuxt":       "Nuxt",
	"nuxtjs":     "NuxtJS",
	"express":    "Express",
	"fastapi":    "FastAPI",
	"django":     "Django",
	"flask":      "Flask",
	"rails":      "Rails",
	"spring":     "Spring",
	"springboot": "Spring Boot",
	"spring boot": "Spring Boot",
	"nodejs":     "Node.js",
	"node.js":    "Node.js",
	"deno":       "Deno",
	"bun":        "Bun",

	// Cloud & Infrastructure
	"saas":       "SaaS",
	"paas":       "PaaS",
	"iaas":       "IaaS",
	"vpc":        "VPC",
	"cdn":        "CDN",
	"dns":        "DNS",
	"ssl":        "SSL",
	"tls":        "TLS",
	"tcp":        "TCP",
	"ip":         "IP",
	"cicd":       "CI/CD",
	"ci/cd":      "CI/CD",
	"devops":     "DevOps",
	"devsecops":  "DevSecOps",
	"sre":        "SRE",
	"ec2":        "EC2",
	"s3":         "S3",
	"rds":        "RDS",
	"lambda":     "Lambda",
	"gcp":        "GCP",
	"azure":      "Azure",

	// Data & Analytics
	"sql":        "SQL",
	"nosql":      "NoSQL",
	"etl":        "ETL",
	"elt":        "ELT",
	"bi":         "BI",
	"kpi":        "KPI",
	"kpis":       "KPIs",
	"roi":        "ROI",

	// Security & Identity
	"iam":        "IAM",
	"rbac":       "RBAC",
	"abac":       "ABAC",
	"acl":        "ACL",
	"pam":        "PAM",
	"iga":        "IGA",
	"siem":       "SIEM",
	"soar":       "SOAR",
	"xdr":        "XDR",
	"edr":        "EDR",
	"cve":        "CVE",
	"owasp":      "OWASP",

	// Month Names
	"january":   "January",
	"february":  "February",
	"march":     "March",
	"april":     "April",
	"may":       "May",
	"june":      "June",
	"july":      "July",
	"august":    "August",
	"september": "September",
	"october":   "October",
	"november":  "November",
	"december":  "December",

	// Day Names
	"monday":    "Monday",
	"tuesday":   "Tuesday",
	"wednesday": "Wednesday",
	"thursday":  "Thursday",
	"friday":    "Friday",
	"saturday":  "Saturday",
	"sunday":    "Sunday",

	// Other Tech Terms
	"io":         "I/O",
	"cpu":        "CPU",
	"gpu":        "GPU",
	"ram":        "RAM",
	"ssd":        "SSD",
	"hdd":        "HDD",
	"os":         "OS",
	"vm":         "VM",
	"vms":        "VMs",
	"pdf":        "PDF",
	"csv":        "CSV",
	"svg":        "SVG",
	"png":        "PNG",
	"jpg":        "JPG",
	"jpeg":       "JPEG",
	"gif":        "GIF",
	"mp3":        "MP3",
	"mp4":        "MP4",
	"wav":        "WAV",
	"webp":       "WebP",
	"webm":       "WebM",
}
