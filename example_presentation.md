---
marp: true
theme: default
paginate: true
style: |
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;600;700&family=JetBrains+Mono:wght@400;500&display=swap');

  section {
    background: linear-gradient(135deg, #0f0c29 0%, #302b63 50%, #24243e 100%);
    color: #e8eaf6;
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    font-size: 24px;
    padding: 40px 60px;
    padding-top: 35px;
  }

  section * {
    border-top-color: transparent !important;
    border-bottom-color: transparent !important;
  }

  /* Definition lists - this is what causes the horizontal lines */
  section dt, section dd {
    border: none !important;
    border-top: none !important;
    border-bottom: none !important;
    padding-bottom: 0 !important;
    margin-bottom: 0 !important;
  }

  section dt::after, section dd::before {
    border: none !important;
    content: none !important;
  }

  section dl {
    border: none !important;
    margin: 0;
    padding: 0;
  }

  section dl::before, section dl::after {
    border: none !important;
    content: none !important;
  }

  /* Target the actual elements more aggressively */
  dt, dd, dl {
    border: 0 !important;
    border-top-width: 0 !important;
    border-bottom-width: 0 !important;
    outline: none !important;
  }

  section::after {
    content: attr(data-marpit-pagination) ' / ' attr(data-marpit-pagination-total);
    position: absolute;
    bottom: 20px;
    right: 60px;
    font-size: 14px;
    color: #7c4dff;
    font-weight: 600;
  }

  section h1 {
    font-size: 47px;
    font-weight: 700;
    margin-top: 0;
    margin-bottom: 0.3em;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    line-height: 1.2;
    border: none;
    border-top: none;
    border-bottom: none;
  }

  section h2 {
    font-size: 36px;
    font-weight: 600;
    color: #b39ddb;
    margin-top: 0.5em;
    margin-bottom: 0.4em;
    border: none;
  }

  section h3 {
    font-size: 28px;
    font-weight: 600;
    color: #9575cd;
    border: none;
  }

  section p, section li {
    font-size: 22px;
    line-height: 1.6;
    color: #e8eaf6;
  }

  section strong {
    color: #7c4dff;
    font-weight: 600;
  }

  section em {
    color: #b39ddb;
    font-style: italic;
  }

  section a {
    color: #7c4dff;
    text-decoration: none;
    border-bottom: 2px solid #7c4dff;
  }

  section code {
    font-family: 'JetBrains Mono', 'Courier New', monospace;
    font-size: 20px;
    background: rgba(124, 77, 255, 0.15);
    color: #b39ddb;
    padding: 2px 8px;
    border-radius: 4px;
    border: 1px solid rgba(124, 77, 255, 0.3);
  }

  section pre {
    font-family: 'JetBrains Mono', 'Courier New', monospace;
    font-size: 18px;
    background: rgba(15, 12, 41, 0.6);
    border: 1px solid #7c4dff;
    border-radius: 8px;
    padding: 20px;
    margin: 1em 0;
    box-shadow: 0 4px 20px rgba(124, 77, 255, 0.2);
  }

  section pre code {
    background: transparent;
    border: none;
    padding: 0;
    color: #f5f5f5;
  }

  section pre code .hljs-string,
  section pre code .hljs-attr {
    color: #e8eaf6;
  }

  section ul, section ol {
    margin: 0.5em 0;
    border: none;
    border-top: none;
    border-bottom: none;
  }

  section li {
    margin: 0.27em 0;
    line-height: 1.44;
    font-size: 20px;
    border: none;
  }

  section table {
    border-collapse: collapse;
    margin: 1em auto;
    background: rgba(15, 12, 41, 0.6);
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid rgba(124, 77, 255, 0.3);
  }

  section th {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: #ffffff;
    font-weight: 600;
    padding: 12px 20px;
    text-align: left;
  }

  section td {
    padding: 10px 20px;
    border-bottom: 1px solid rgba(124, 77, 255, 0.2);
    color: #e8eaf6;
    background: rgba(15, 12, 41, 0.4);
  }

  section tr:last-child td {
    border-bottom: none;
  }

  section tr:hover td {
    background: rgba(124, 77, 255, 0.15);
  }

  section blockquote {
    border-left: 4px solid #7c4dff;
    background: rgba(124, 77, 255, 0.1);
    padding: 12px 20px;
    margin: 1em 0;
    border-radius: 0 8px 8px 0;
  }

  /* Title slide styling */
  section:first-of-type h1 {
    font-size: 58px;
    text-align: center;
    margin-top: 1em;
  }

  section:first-of-type h2 {
    text-align: center;
    font-size: 32px;
    color: #b39ddb;
  }

  section:first-of-type p {
    text-align: center;
    font-size: 22px;
    color: #9575cd;
  }

  /* Emoji and icon styling */
  section img[alt~="emoji"], section img[alt~="icon"] {
    height: 1.2em;
    vertical-align: middle;
  }

  /* Hide any horizontal rules that Marp might generate */
  section hr {
    display: none;
  }

  /* VibeMinds.AI branding footer */
  section::before {
    content: 'VibeMinds.AI';
    position: absolute;
    bottom: 20px;
    left: 60px;
    font-size: 12px;
    color: #7c4dff;
    font-weight: 600;
    letter-spacing: 0.5px;
  }
---

<!--
Welcome to our presentation on the Statistics Agent Team project. Today, we'll explore how we built a sophisticated multi-agent system for finding and verifying statistics from the web using Go and large language models.
[PAUSE:1000]
This project was born out of a fundamental problem: how can we trust the statistics that AI systems give us? As we built this system, we encountered numerous technical challenges, from L L M hallucinations to architectural decisions about security and scalability. We'll walk through each challenge and show you how we solved them.
[PAUSE:1500]
-->

# Statistics Agent Team
## Building a Multi-Agent System for Verified Statistics

**A Production-Ready Implementation**

Built with Google ADK, Eino, and Multi-LLM Support

---

<!--
Let's start by understanding what problem we're trying to solve. When you ask a chatbot for statistics, how do you know if the numbers are accurate? Can you verify the source? This is the core challenge we addressed.
[PAUSE:1500]
The fundamental issue is that large language models are trained on historical data and often hallucinate statistics. They'll confidently give you numbers that sound plausible but are completely fabricated. Even worse, they'll generate URLs that look legitimate but lead to pages that don't exist or have moved. This creates a credibility crisis. How can researchers, journalists, or analysts trust A I generated statistics when there's no way to verify them?
[PAUSE:2000]
-->

# The Problem 🎯

**Challenge**: Finding verified, numerical statistics from reputable web sources

**Pain Points** —
- ❌ LLMs hallucinate statistics and sources
- ❌ URLs from LLM memory are often outdated or wrong
- ❌ No verification that excerpts actually exist
- ❌ Hard to distinguish reputable vs unreliable sources

**Goal**: Build a system that provides **provably accurate** statistics

---

<!--
We established clear requirements for what success looks like. The system must not only find statistics but verify them against actual web sources. Speed matters, but accuracy is paramount.
[PAUSE:1500]
The key challenge here was balancing comprehensiveness with performance. We needed to search enough sources to find diverse statistics, extract them intelligently without missing any, and verify each one rigorously, all while keeping response times under sixty seconds. Setting these concrete targets, especially the sixty to ninety percent verification rate goal, gave us a clear benchmark to measure against. Most importantly, we insisted on supporting multiple L L M providers from the start, because we knew different organizations have different preferences and constraints.
[PAUSE:2000]
-->

# Requirements 📋

## Functional Requirements
- ✅ Search web for statistics on any topic
- ✅ Extract numerical values with context
- ✅ Verify excerpts exist in source documents
- ✅ Validate numerical accuracy
- ✅ Prioritize reputable sources (.gov, .edu, research orgs)

## Non-Functional Requirements
- ✅ 60-90% verification rate (vs 0% for direct LLM)
- ✅ Response time: under 60 seconds
- ✅ Support multiple LLM providers
- ✅ Containerized deployment

---

<!--
We chose a four-agent architecture with clear separation of concerns. Each agent has a specific responsibility in the pipeline. This modular design allows us to optimize each component independently.
[PAUSE:1500]
The architecture decision was crucial. We could have built a monolithic system where one L L M does everything, but we learned early on that this doesn't work. Different tasks need different capabilities. Search needs to be fast and comprehensive, but doesn't need an L L M at all. Extraction needs sophisticated language understanding. Verification needs to be rigorous and deterministic. By separating these concerns, we could optimize each agent independently, swap out implementations, and debug issues more easily. This modular approach also meant multiple developers could work in parallel without stepping on each other's toes.
[PAUSE:2500]
-->

# Architecture Overview 🏗️

```
User Request → Orchestrator
                    ↓
        ┌───────────┼───────────┐
        ↓           ↓           ↓
    Research    Synthesis  Verification
    (Search)    (Extract)   (Validate)
        ↓           ↓           ↓
      URLs    Statistics    Verified ✓
```

**4 Specialized Agents** working together:
1. **Research** - Web search (no LLM needed)
2. **Synthesis** - LLM-based extraction
3. **Verification** - Web validation
4. **Orchestration** - Workflow coordination

---

<!--
The research agent is our foundation. It performs web search using Google via Serper or SerpAPI. Notice it doesn't use an LLM at all - it's pure search functionality. This keeps it fast and cost-effective.
[PAUSE:2500]
-->

# Agent 1: Research Agent 🔍

**Responsibility**: Find relevant web sources

**Implementation** —
- No LLM required (pure search)
- Integrates with Serper/SerpAPI via `metasearch` library
- Filters for reputable domains
- Returns 30 URLs by default

**Key Decision**: Separate search from extraction
- Allows caching of search results
- Different providers don't need LLM changes
- Faster iteration on search queries

**Port**: 8001

---

<!--
The synthesis agent is where the heavy lifting happens. It fetches actual web pages and uses an L L M to intelligently extract statistics. We built this with Google's A D K framework for robust L L M operations.
[PAUSE:1500]
This agent turned out to be the most challenging to get right. The problem is that web pages are messy. They have navigation, ads, and irrelevant content mixed with the statistics we want. We needed to extract just the numerical data with enough context to understand what it means, while preserving the exact wording so we can verify it later. We also had to handle different page structures, formats, and writing styles. The L L M is perfect for this kind of intelligent extraction, but we had to carefully tune how much content to give it and how many pages to process to get good coverage.
[PAUSE:2500]
-->

# Agent 2: Synthesis Agent 📊

**Responsibility**: Extract statistics from web pages

**Implementation** (Google ADK):
- Fetches webpage content (30K chars per page)
- LLM analyzes text for numerical statistics
- Extracts verbatim excerpts
- Processes 15+ pages for comprehensive coverage
- Returns candidates with metadata

**Key Challenge**: Getting complete extraction
- ❌ Initial: Only returned 5-8 statistics
- ✅ Solution: Increased pages (5→15), content (15K→30K), multiplier (2x→5x)

**Port**: 8004

---

<!--
Here's a critical insight we learned. When we first tested, the synthesis agent would only find a handful of statistics. We discovered we needed to cast a much wider net because many candidates fail verification. The five-x multiplier accounts for this reality.
[PAUSE:1500]
This was probably our biggest "aha moment" during development. Our initial version processed only five pages with fifteen thousand characters each, and used a conservative two-x multiplier. We'd get back maybe five to eight statistics. Meanwhile, Chat G P T dot com was returning twenty plus statistics on the same query. We were confused at first. Was their L L M just better? The real issue was that many statistics fail verification. Pages move. Content changes. Excerpts don't match exactly. So even if you extract fifty candidate statistics, only thirty percent might verify successfully. That's why we needed to be aggressive upfront, processing fifteen plus pages with thirty thousand characters each, and using a five-x multiplier. Cast a wide net, then filter rigorously.
[PAUSE:3000]
-->

# Synthesis Agent: Key Learnings 💡

**Problem**: Low statistical yield (5-8 stats vs ChatGPT's 20+)

**Root Cause Analysis** —
- Too few pages processed (only 5)
- Too little content per page (15K chars)
- Too conservative multiplier (2x)

**Solution** - Aggressive extraction:
```go
minPagesToProcess := 15  // (increased from 5)
maxContentLen := 30000   // (increased from 15K)
multiplier := 5          // (increased from 2x)
```

**Result**: Now matches ChatGPT.com performance! 🎉

---

<!--
The verification agent is what sets our system apart. It doesn't just trust what the L L M says, it actually fetches the source U R L and validates the excerpt exists word for word. This is where accuracy comes from.
[PAUSE:1500]
This is the trust layer of our system. When the synthesis agent extracts a statistic and says "I found this on this U R L," we don't just take its word for it. We independently fetch that exact U R L again, extract the text, and search for the claimed excerpt. If we find it verbatim, great. If not, we use a light L L M check for fuzzy matching to handle minor formatting differences. But fundamentally, we're validating that the statistic actually exists in the source. This catches L L M hallucinations, pages that have changed, paywalls, broken links, and all sorts of other real-world issues. It's more expensive and slower, but it's the only way to achieve real accuracy.
[PAUSE:2500]
-->

# Agent 3: Verification Agent ✅

**Responsibility**: Validate statistics against sources

**Implementation** (Google ADK):
- Re-fetches source URLs
- Checks excerpts exist verbatim
- Validates numerical values match exactly
- Uses light LLM assistance for fuzzy matching
- Returns pass/fail with detailed reasons

**Key Decision**: Always fetch original source
- No trusting LLM claims
- Catches hallucinations
- Verifies pages haven't changed

**Port**: 8002

---

<!--
We implemented two orchestration approaches. The A D K version uses an L L M to make decisions about workflow. The Eino version uses a deterministic graph. Both work, but Eino is faster and more predictable for production use.
[PAUSE:1500]
This was an interesting architectural choice. Initially we used Google A D K for orchestration, where an L L M decides what to do next. It's flexible and adaptive but has problems. The L L M might make different decisions on the same input, leading to non-deterministic behavior. It's slower because every decision requires an L L M call. And it's harder to debug because the L L M's reasoning isn't always clear. So we built a second implementation using Eino, with a deterministic directed graph. Every query follows the same path: validate, research, synthesize, verify, format. It's faster, cheaper, reproducible, and much easier to reason about. For production systems, deterministic behavior is usually better than adaptive flexibility.
[PAUSE:2500]
-->

# Agent 4: Orchestration Agent 🎭

**Two Implementations Available** —

## Option A: Google ADK (LLM-driven)
- Uses LLM to decide workflow steps
- Adaptive retry logic
- More flexible but slower

## Option B: Eino (Deterministic) ⭐ **RECOMMENDED**
- Type-safe graph-based workflow
- Predictable, reproducible behavior
- Faster and lower cost
- No LLM for orchestration decisions

**Both run on Port 8000** (choose one)

---

<!--
Here's what the Eino workflow looks like. It's a directed graph where each node has a specific job. Data flows predictably from input validation through to formatted output. This determinism is crucial for production reliability.
[PAUSE:2500]
-->

# Eino Orchestration Flow 🔄

```
ValidateInput
     ↓
Research (30 URLs)
     ↓
Synthesis (15+ pages → candidates)
     ↓
Verification (validate each)
     ↓
QualityCheck (≥ min verified?)
     ↓
FormatOutput → User
```

**Why Eino?**
- Type-safe operations
- No non-deterministic LLM decisions
- Easier to debug and test
- Production-ready reliability

---

<!--
One of our biggest challenges was the direct mode. Initially, we thought letting an L L M directly answer from memory would be useful. What we found was eye opening: zero percent verification rate. This taught us the importance of real-time web search.
[PAUSE:1500]
This was a humbling discovery. We built a direct mode where the L L M answers from its training data, just like asking Chat G P T without web search. The L L M would confidently return ten statistics with seemingly legitimate U R Ls. But when we ran them through verification, literally zero verified. The problem? The L L M was guessing U R Ls based on patterns it learned during training. It would say "according to this N I H study" and generate a plausible looking N I H dot gov U R L, but that specific page didn't exist, had moved, or never contained that statistic. The L L M's training data was up to January twenty twenty-five, but web pages change constantly. This was our first major lesson: for real-time factual information, L L M memory is not enough. You absolutely need live web search.
[PAUSE:3000]
-->

# Challenge 1: Direct Mode Failure ⚠️

**Initial Idea**: Let LLM answer from memory (like ChatGPT)

**Implementation** —
```bash
./stats-agent search "AI trends" --direct
```

**The Problem** —
- LLM returns statistics from training data (up to Jan 2025)
- URLs are **guessed** - not from real search
- Pages have moved, changed, or are paywalled
- **0% verification rate** when validated

**The Lesson**: Real-time web search is essential for statistics

---

<!--
Here's a real comparison we did. We asked both systems about the same topic. Chat G P T dot com returned many verifiable statistics because it uses real-time Bing search. Our direct mode returned plausible looking numbers but with completely wrong U R Ls. This comparison drove our architecture decisions.
[PAUSE:1500]
This table tells the whole story. When we tested Chat G P T dot com, the web version, not the A P I, on the query "A I trends," it returned over twenty statistics, and ninety percent of them verified. Our direct mode returned ten statistics, and zero percent verified. Initially we thought maybe Open A I's L L M was just better than Gemini. But that wasn't it. The key insight is in the "why" column. Chat G P T dot com's success comes from real-time Bing search integration, not from having a better language model. It searches the web live, fetches current pages, and extracts statistics from actual sources. That's exactly what we needed to do. So we built pipeline mode with Serper and Serp A P I integration for real-time Google search, and immediately our verification rate jumped to sixty to ninety percent. The lesson: architecture matters more than model quality for this use case.
[PAUSE:3000]
-->

# Direct Mode vs ChatGPT.com 📊

**Same Query: "AI trends"**

| System | Statistics Found | Verification Rate | Why? |
|--------|-----------------|-------------------|------|
| **ChatGPT.com** | 20+ | ✅ 90%+ | Real-time Bing search |
| **Direct Mode** | 10 | ❌ 0% | LLM memory (outdated URLs) |
| **Pipeline Mode** | 15-25 | ✅ 60-90% | Real-time Google search |

**Key Insight**: ChatGPT.com's success comes from **web search**, not just LLM quality!

**Our Solution**: Pipeline mode with Serper/SerpAPI

---

<!--
Based on this learning, we added clear warnings to our documentation. Direct mode remains available for general knowledge questions, but we steer users toward pipeline mode for actual statistics. Being honest about limitations builds trust.
[PAUSE:2500]
-->

# Solution: Pipeline Mode ✅

**What We Changed** —
- Made Pipeline mode the default
- Added warnings to Direct mode docs
- Implemented hybrid mode (Direct + Verification)

**README Warning** —
```markdown
⚠️ Direct Mode - Not Recommended for Statistics
- ❌ Uses LLM memory (training data)
- ❌ Outdated URLs
- ❌ 0% verification rate

✅ For statistics, use Pipeline mode instead
```

**Result**: Clear expectations, better user experience

---

<!--
The second major challenge was L L M provider flexibility. Different teams use different L L M vendors. We needed to support them all without duplicating code. The solution was a factory pattern with provider abstraction.
[PAUSE:1500]
This challenge emerged from real-world requirements. Some organizations are all in on Google Gemini. Others prefer Anthropic Claude for its reasoning capabilities. Some teams want Open A I for familiarity. Others need to use local models via Ollama for privacy or cost reasons. And then there's X A I Grok for those who want cutting edge performance. Each provider has completely different A P Is, authentication methods, model names, rate limits, and pricing. We could have just picked one and stuck with it, but that would limit adoption. Instead, we needed a flexible architecture that abstracts away these differences, so agents don't care which L L M they're using. The challenge was building this abstraction without sacrificing provider-specific features or performance.
[PAUSE:2500]
-->

# Challenge 2: Multi-LLM Support 🔧

**Requirement**: Support multiple LLM providers

**Supported Providers** —
- Google Gemini (default) - `gemini-2.5-flash` / `gemini-2.5-pro`
- Anthropic Claude - `claude-sonnet-4-20250514` / `claude-opus-4-1-20250805`
- OpenAI - `gpt-4o` / `gpt-5`
- xAI Grok - `grok-4-1-fast-reasoning` / `grok-4-1-fast-non-reasoning`
- Ollama - `llama3:8b` / `mistral:7b` (local)

**Challenge**: Each provider has different APIs, models, rate limits

**Solution**: Abstraction via `gollm` library

---

<!--
Here's how the abstraction works. The gollm library provides a unified interface. We just select a provider via environment variable. The agents don't care which L L M they're using, they just call the standard interface.
[PAUSE:1500]
The factory pattern was key to solving this cleanly. We created a create L L M function that takes a config object and returns a generic client interface. Inside, it switches on the L L M provider string and calls the appropriate provider-specific creation function. Each function handles that provider's quirks: Gemini needs a Google A P I key, Claude needs an Anthropic key, Ollama needs a local U R L and doesn't need an A P I key at all. But they all return the same interface, so the synthesis and verification agents can use any provider without changing their code. Want to test Claude versus Gemini? Just change one environment variable. This flexibility made development much faster and enabled users to choose based on their constraints.
[PAUSE:2500]
-->

# Multi-LLM Implementation 🎯

**Factory Pattern** in `pkg/llm/factory.go`:

```go
func CreateLLM(cfg *config.Config) (*genai.Client, string, error) {
    switch cfg.LLMProvider {
    case "gemini":
        return createGeminiClient(cfg)
    case "claude":
        return createClaudeClient(cfg)
    case "openai":
        return createOpenAIClient(cfg)
    case "xai":
        return createXAIClient(cfg)
    case "ollama":
        return createOllamaClient(cfg)
    }
}
```

**Benefit**: Agents are provider-agnostic

---

<!--
Configuration is entirely environment-based. No hardcoded API keys. This makes it secure and flexible. You can switch providers with a single environment variable change. Perfect for testing different models or working around rate limits.
[PAUSE:2500]
-->

# LLM Configuration Example 💻

**Simple Environment Variables** —

```bash
# Use Gemini (default)
export GOOGLE_API_KEY="your-key"

# Switch to Claude
export LLM_PROVIDER="claude"
export ANTHROPIC_API_KEY="your-key"

# Switch to local Ollama
export LLM_PROVIDER="ollama"
export OLLAMA_URL="http://localhost:11434"
export LLM_MODEL="llama3:8b"
```

**No code changes required!** 🎉

---

<!--
The third challenge was search integration. Web search A P Is are not free, and different organizations prefer different providers. We needed flexibility here too. The metasearch library provided the abstraction we needed.
[PAUSE:1500]
This was similar to the L L M challenge but for search. Web search A P Is cost money. Serper costs fifty dollars a month for five thousand queries. Serp A P I has different pricing tiers. Some teams already have contracts with specific providers. Others want to use mock data during development to avoid A P I costs. Each search A P I returns results in different formats, with different fields and structures. We needed the same kind of abstraction we built for L L Ms. The metasearch library solves this by providing a unified search interface. You call search normalized, and it returns a standard result format regardless of which provider is actually doing the search. This means the research agent doesn't need to know or care whether it's using Serper, Serp A P I, or a mock provider. Flexibility without complexity.
[PAUSE:2500]
-->

# Challenge 3: Search Provider Options 🔍

**Requirement**: Support multiple search providers

**Options** —
- **Serper API** - $50/month, 5K queries (recommended)
- **SerpAPI** - Alternative with different pricing
- **Mock** - For development without API keys

**Challenge**: Different APIs, different response formats

**Solution**: `metasearch` library abstraction

```go
// Unified interface - works with any provider
result, err := searchClient.SearchNormalized(ctx, params)
```

---

<!--
Early on, we made a security mistake. The direct mode ran the L L M on the client side. This meant users needed A P I keys. Not only is this a security risk, but it's inconvenient. We moved to a server-side architecture.
[PAUSE:1500]
This was an architectural flaw we caught relatively early. In our first version of direct mode, the client C L I tool would load the A P I key from the user's environment and make L L M calls directly. This is bad for several reasons. First, every user needs their own A P I key, which is friction for adoption. Second, A P I keys in client environments can leak. Third, you can't update the prompts without users pulling new code. Fourth, there's no centralized rate limiting or cost control. It's a distributed mess. The fix was to create a direct agent server that runs on port eight zero zero five. Now clients make H T T P requests to the server, and the server holds the A P I keys securely. You can update prompts server side, implement rate limiting, monitor costs, and users don't need any credentials. It's the right architecture for production.
[PAUSE:2500]
-->

# Challenge 4: Security Architecture 🔒

**Initial Design**: Client-side LLM (❌ Bad)
```bash
# Client needs API key!
export GOOGLE_API_KEY="key"
./stats-agent search "topic" --direct
```

**Problem** —
- Clients need API keys (security risk)
- Hard to update prompts
- No centralized rate limiting

**Solution**: Server-side Direct Agent (✅ Good)
- Direct Agent server on port 8005
- Client makes HTTP requests
- Server holds API keys
- Centralized control

---

<!--
The server-side architecture also gave us an opportunity to add proper API documentation. We used the Huma framework to generate OpenAPI three-point-one specs automatically. Now external clients can easily integrate with interactive Swagger docs.
[PAUSE:2500]
-->

# Direct Agent Server Implementation 🌐

**Built with Huma v2 + Chi router** —
- OpenAPI 3.1 automatic generation
- Interactive Swagger UI at `/docs`
- Type-safe request/response handling
- Proper HTTP timeouts

**Example** —
```go
type DirectSearchInput struct {
    Body struct {
        Topic    string `json:"topic" minLength:"1"`
        MinStats int    `json:"min_stats" minimum:"1"`
    }
}

huma.Register(api, operation, handler)
```

**Port 8005** - Production-ready with docs! 📚

---

<!--
One subtle but important challenge was number formatting. J S O N doesn't allow commas in numbers. But L L Ms love to format numbers like humans do, with commas. This caused silent parsing failures until we fixed the prompts.
[PAUSE:1500]
This was one of those bugs that took way too long to find. The L L M would return what looked like perfectly valid J S O N. The structure was right, all the fields were there, but our J S O N parser would fail with a cryptic error. When we finally inspected the raw L L M output carefully, we found the issue: numbers like two thousand five hundred thirty seven were being written as two comma five three seven. To humans, that's correct formatting. But in J S O N, numbers cannot have commas. It's syntactically invalid. The L L M was being helpful by formatting numbers the way humans expect, but breaking the J S O N spec. The fix was to add very explicit instructions in the prompt: the value field must be a plain number with no commas. We even gave examples. After that, no more parsing errors. This taught us that L L Ms need very explicit formatting instructions, especially for structured output.
[PAUSE:2500]
-->

# Challenge 5: JSON Number Format 🔢

**The Bug** —
```json
{
  "value": 2,537  // ❌ Invalid JSON!
}
```

**Root Cause**: LLM formats numbers like humans (2,537)

**The Fix** - Explicit prompt instructions:
```
CRITICAL: The "value" field must be a plain number
with NO commas (e.g., 2537 not 2,537)

REMEMBER: Numbers like 75,000 should be written
as 75000 (no comma).
```

**Result**: Valid JSON every time! ✅

---

<!--
We also discovered the importance of explicit completeness instructions. L L Ms tend to be lazy. They'll find one or two examples and stop. We had to explicitly tell them to find all statistics on a page, not just a few examples.
[PAUSE:1500]
This was a frustrating pattern we kept seeing. We'd feed the synthesis agent a page about climate change that clearly had ten different statistics scattered throughout. The L L M would extract maybe one or two and call it done. Why? Because L L Ms are trained to be concise and helpful. If you ask for statistics, they assume you want a few representative examples, not an exhaustive list. They're being efficient from their perspective. But we needed completeness. So we had to be extremely explicit in the prompts. We added phrases like "extract every statistic you find, not just one or two," and "if the page contains ten statistics, return ten items in the array." We told it to only return an empty array if absolutely no statistics are found. This kind of explicit instruction made a huge difference, increasing extraction by two to three x per page. The lesson: L L Ms don't read your mind. Be ridiculously explicit about what complete means.
[PAUSE:2500]
-->

# Prompt Engineering Lessons 📝

**Problem**: LLM returns 1-2 statistics, stops

**Bad Prompt** —
```
Find statistics about climate change.
```

**Good Prompt** —
```
Extract EVERY statistic you find, not just one or two.
Be thorough and comprehensive.

If the page contains 10 statistics, return 10 items
in the array.

Return empty array [] ONLY if absolutely no statistics
are found.
```

**Impact**: 2-3x more statistics extracted per page

---

<!--
Deployment was a key consideration. We needed to support both local development and production Docker deployments. Make commands handle local, Docker Compose handles production. Both use the same code and configuration.
[PAUSE:1500]
Developer experience matters. You need to be able to run locally for development, but deploy to production easily. We support both with the same codebase. For local development, make run all eino starts all four agents in the foreground where you can see logs. Then you run the C L I client to make requests. For production, docker compose up dash d runs all agents as containerized services. They communicate via H T T P on their assigned ports: eight thousand through eight thousand two, and eight thousand four through eight thousand five. The configuration is identical, just environment variables. This seamless transition from local to production means you're testing the real system locally, not some simplified mock. What you develop is what you deploy.
[PAUSE:2500]
-->

# Deployment Architecture 🐳

**Two Deployment Methods** —

## Local Development
```bash
make run-all-eino  # Start all 4 agents
./bin/stats-agent search "topic"
```

## Docker Production
```bash
docker-compose up -d  # All agents containerized
curl -X POST http://localhost:8000/orchestrate
```

**Same code, same config** - seamless transition!

**Ports**: 8000-8002, 8004-8005

---

<!--
We also added an M C P server for integration with Claude Code and other A I tools. This allows our statistics engine to be used as a tool by other A I agents. It's a nice example of composability in multi-agent systems.
[PAUSE:1500]
This is where things get meta. Model Context Protocol, or M C P, is a standard for exposing tools to A I assistants. We implemented an M C P server that wraps our statistics system, making it available to Claude Code, the A I coding assistant. Now, when you're working in Claude Code and ask "find me statistics about renewable energy adoption," Claude Code can call our M C P server, which triggers the full pipeline, searches the web, verifies statistics, and returns results. Claude Code then incorporates those verified statistics into your code or documentation. It's composability at the A I agent level. Our agent team becomes a tool for other agents. This pattern of exposing capabilities via standard protocols is crucial for building ecosystems of specialized agents that work together.
[PAUSE:2500]
-->

# MCP Server Integration 🔌

**Model Context Protocol** support for AI tool integration

**Use Case**: Claude Code can search for verified statistics

```json
{
  "mcpServers": {
    "stats-agent": {
      "command": "go",
      "args": ["run", "mcp/server/main.go"]
    }
  }
}
```

**Tools Available** —
- `search_statistics` - Full pipeline search
- `verify_statistic` - Single verification

**Integration**: Works with Claude Code, other MCP clients

---

<!--
Let's talk results. The pipeline mode achieves sixty to ninety percent verification rate. Response times are under a minute for most queries. Compare this to direct mode's zero percent verification, and you can see why architecture matters.
[PAUSE:1500]
This table summarizes the tradeoffs. Direct mode is fast, five to ten seconds, but has terrible accuracy and zero verification. Pipeline mode takes thirty to sixty seconds, but achieves sixty to ninety percent verification with high accuracy. It searches thirty real U R Ls, processes fifteen plus pages, and validates every statistic. The cost is higher because we're doing real work, making real A P I calls, fetching real pages. But the value is in the verification. If you need actual, verified statistics for a research report or data analysis, pipeline mode is the only choice. If you just want to brainstorm ideas quickly and accuracy doesn't matter, direct mode might be acceptable. The key insight is that there's no free lunch. Accuracy requires work, and work takes time and money.
[PAUSE:2500]
-->

# Performance Metrics 📈

| Metric | Direct Mode | Pipeline Mode |
|--------|-------------|---------------|
| **Verification Rate** | ❌ 0-30% | ✅ 60-90% |
| **Response Time** | ⚡ 5-10s | ⚡ 30-60s |
| **URLs Searched** | 0 (LLM memory) | 30 (real search) |
| **Pages Processed** | 0 | 15+ |
| **Cost per Query** | Low | Medium |
| **Accuracy** | ❌ Low | ✅ High |

**Sweet Spot**: Pipeline mode for statistics, Direct for general Q&A

---

<!--
Here's a concrete example. When we search for climate change statistics, we get back verified data with exact sources. Notice the verbatim excerpt, that's proof it came from the actual source. This is what makes our system trustworthy.
[PAUSE:1500]
This J S O N output shows what a verified statistic looks like. The name field describes what the statistic measures: global temperature increase. The value is one point one with unit degrees Celsius. The source is the I P C C Sixth Assessment Report, a highly reputable climate science organization. The source U R L is the actual page. But most importantly, look at the excerpt field. It contains the verbatim text from that page: "Global surface temperature has increased by approximately one point one degrees Celsius since pre-industrial times." We fetched that page, extracted the text, and found this exact sentence. That's why verified is true. This isn't an L L M guessing or hallucinating. This is real data from a real source, programmatically verified. That's the trust guarantee we provide.
[PAUSE:2500]
-->

# Real-World Example 🌍

**Query**: "climate change statistics"

**Result** —
```json
{
  "name": "Global temperature increase",
  "value": 1.1,
  "unit": "°C",
  "source": "IPCC Sixth Assessment Report",
  "source_url": "https://www.ipcc.ch/...",
  "excerpt": "Global surface temperature has increased
             by approximately 1.1°C since pre-industrial
             times...",
  "verified": true
}
```

**Verification**: Excerpt found verbatim in source! ✅

---

<!--
The technology choices were deliberate. Go provided concurrency and performance. A D K gave us robust L L M operations. Eino provided deterministic orchestration. Together they create a production-ready system.
[PAUSE:1500]
Let's talk about why we chose each technology. Go was chosen for its concurrency model, fast performance, and simple deployment. You get a single binary with no dependencies. Google A D K provides robust L L M operations with built-in retry logic, structured output, and tool calling. It handles the complexity of L L M interactions. Eino provides deterministic graph-based orchestration with type safety and reproducible behavior. Huma v2 generates Open A P I three point one specs automatically, giving us great documentation for free. Chi v5 is a lightweight H T T P router that doesn't get in the way. The gollm library abstracts multiple L L M providers so we're not locked into one vendor. And metasearch does the same for search A P Is. These choices prioritize flexibility, reliability, and developer experience. We could build new features quickly without fighting the tech stack.
[PAUSE:2500]
-->

# Technology Stack 🛠️

**Language & Runtime** —
- Go 1.21+ - Concurrency, performance, simple deployment

**Agent Frameworks** —
- **Google ADK** - LLM-based agent operations
- **Eino** - Deterministic graph orchestration

**API & Docs** —
- **Huma v2** - OpenAPI 3.1 generation
- **Chi v5** - Lightweight HTTP router

**Integrations** —
- **gollm** - Multi-provider LLM abstraction
- **metasearch** - Unified search API

---

<!--
We learned several key lessons building this system. Real-time search beats L L M memory for current data. Verification is non-negotiable for accuracy. Clear separation of concerns makes debugging easier. And always be explicit with L L Ms, they need detailed instructions.
[PAUSE:1500]
These lessons were hard won through trial and error. The zero percent verification rate in direct mode versus sixty to ninety percent in pipeline mode taught us that real-time search is essential. The discovery that many extracted statistics fail verification taught us to always validate against sources. The ability to optimize each agent independently taught us the value of modularity. The J S O N parsing failures and incomplete extractions taught us that prompt engineering is critical, not optional. And the need to support multiple L L M providers and search providers taught us that flexibility drives adoption. These aren't just technical lessons, they're architectural principles that apply to any multi-agent system. Get the architecture right, and the implementation follows. Get it wrong, and you'll fight issues forever.
[PAUSE:3000]
-->

# Key Learnings 💡

1. **Real-time search > LLM memory** for current data
   - 0% vs 60-90% verification rate
2. **Verification is non-negotiable** for accuracy
   - Always fetch and validate sources
3. **Separation of concerns** enables optimization
   - Search, extract, verify are independent
4. **Prompt engineering matters** at scale
   - Explicit completeness instructions needed
5. **Flexibility enables adoption**
   - Multi-LLM, multi-search provider support

---

<!--
Some challenges remain. Paywalled content is inaccessible. Different languages need special handling. And we'd love to support statistical ranges, not just single values. These are areas for future enhancement.
[PAUSE:1500]
No system is perfect, and ours has known limitations. Paywalled content behind subscriptions like the New York Times or academic journals is inaccessible without credentials. We can see the page exists, but can't fetch the content. Non-English sources require translation layers. And range statistics like "seventy nine to ninety six percent" don't fit our current schema that expects a single value field. These aren't blockers, but they limit coverage. On the roadmap, we're planning to add a value max field for ranges, integrate Perplexity A P I which has built-in search, add caching to avoid redundant searches, implement streaming for better perceived performance, and add multi-language support. The foundation is solid, now it's about expanding capabilities based on user feedback.
[PAUSE:2500]
-->

# Challenges & Future Work 🚀

**Current Limitations** —
- ❌ Paywalled content inaccessible
- ❌ Non-English sources need translation
- ⚠️ Range statistics (e.g., "79-96%") need schema updates

**Future Enhancements** —
- ✨ Add `value_max` field for ranges
- ✨ Perplexity API integration (built-in search)
- ✨ Caching layer for search results
- ✨ Streaming responses for faster perceived performance
- ✨ Multi-language support

---

<!--
Here's the complete workflow from user query to verified results. Each step is optimized and reliable. The human in the loop retry gives users control when results are partial. This balance of automation and control is key.
[PAUSE:1500]
Let's walk through a concrete example. The user runs stats dash agent search "renewable energy" with a minimum of ten statistics. Here's what happens behind the scenes. First, the orchestrator validates the input. Is the topic non-empty? Is ten a reasonable target? Then it calls the research agent, which searches thirty U R Ls via Serper. Next, the synthesis agent processes fifteen plus pages, extracting over four hundred fifty thousand characters of total content. It uses the L L M to extract fifty plus candidate statistics from this corpus. Then comes the critical verification stage. Each candidate is independently validated. Out of fifty candidates, twelve verify successfully, which is a sixty percent verification rate. The orchestrator checks: twelve is greater than or equal to ten, so the quality threshold is met. Finally, it formats the output as J S O N and returns it to the user. Total time: around forty five seconds. The entire process is logged, observable, and reproducible.
[PAUSE:2500]
-->

# Complete Workflow Example 🔄

```bash
./stats-agent search "renewable energy" --min-stats 10
```

**What Happens** —
1. **Orchestrator** validates input
2. **Research** searches 30 URLs via Serper
3. **Synthesis** processes 15+ pages (450K+ chars total)
4. **Synthesis** extracts 50+ candidate statistics
5. **Verification** validates each candidate
6. **Verification** returns 12 verified (60% rate)
7. **Orchestrator** checks: 12 ≥ 10 ✅
8. **User** receives JSON output

**Total time**: ~45 seconds

---

<!--
Monitoring and observability were important. Each agent logs its operations. We can see how many pages were processed, how many candidates were extracted, and the verification pass rate. This helps us continually optimize the system.
[PAUSE:2500]
-->

# Monitoring & Observability 📊

**Structured Logging** at each stage:

```
Research Agent: Found 30 search results
Synthesis Agent: Extracted 8 statistics from nature.com
Synthesis Agent: Total candidates: 52 from 15 pages
Verification Agent: Verified 10/15 candidates (67%)
Orchestration: Target met (10 verified)
```

**Health Checks** —
- `/health` endpoint on each agent
- Docker health checks in production
- Timeout monitoring (60s max)

**Metrics to Track** —
- Verification rate per query
- Average response time
- Cost per query (API calls)

---

<!--
Make commands provide a simple interface for complex operations. Developers can start the entire system with one command. This developer experience was a priority - if it's hard to run locally, it won't get used.
[PAUSE:2500]
-->

# Developer Experience 👨‍💻

**Simple Commands** —

```bash
# Install dependencies
make install

# Build all agents
make build

# Run everything (Eino orchestrator)
make run-all-eino

# Run direct + verification only
make run-direct-verify

# Run tests
make test
```

**Clean Abstractions**: Agents don't know about each other's internals

**Easy Debugging**: Run individual agents in separate terminals

---

<!--
Configuration is centralized but flexible. The dot-env file approach means you can have different environments easily. Development, staging, and production configs are just different env files. No code changes needed.
[PAUSE:2500]
-->

# Configuration Management ⚙️

**Environment-Based** —

```bash
# .env file
LLM_PROVIDER=gemini
GOOGLE_API_KEY=your-key
SEARCH_PROVIDER=serper
SERPER_API_KEY=your-key
```

**Override per Agent** —
```bash
# Use different LLM for synthesis
export SYNTHESIS_LLM_PROVIDER=claude
export SYNTHESIS_LLM_MODEL=claude-sonnet-4-20250514
```

**Docker-Friendly**: All config via environment variables

---

<!--
Let's compare the three operating modes side by side. Each has a use case. Direct mode is for brainstorming when you don't need verification. Hybrid adds verification but suffers from the LLM memory problem. Pipeline mode is the gold standard for actual statistics.
[PAUSE:3000]
-->

# Mode Comparison Summary 📊

| Feature | Direct | Hybrid | Pipeline |
|---------|--------|---------|----------|
| **Speed** | ⚡⚡⚡ 5s | ⚡⚡ 15s | ⚡ 45s |
| **Accuracy** | ❌ Low | ⚠️ Medium | ✅ High |
| **Verification** | ❌ No | ⚠️ LLM URLs | ✅ Real URLs |
| **Cost** | $ | $$ | $$$ |
| **Use Case** | Brainstorm | Quick check | Production |
| **Agents Needed** | 1 | 2 | 4 |

**Recommendation**: Pipeline mode for statistics that matter

---

<!--
Testing was multi-layered. Unit tests for individual functions. Integration tests for agent communication. End-to-end tests for complete workflows. And manual testing against known statistics to verify accuracy.
[PAUSE:2500]
-->

# Testing Strategy 🧪

**Unit Tests** —
- Individual function validation
- LLM provider factory
- JSON parsing edge cases

**Integration Tests** —
- Agent-to-agent communication
- HTTP endpoint validation
- Error handling flows

**End-to-End Tests** —
- Complete pipeline execution
- Verification rate validation
- Performance benchmarks

**Manual Testing** —
- Known statistics verification
- Multi-provider compatibility
- Edge case exploration

---

<!--
Error handling was crucial for reliability. Network failures happen. Sources go offline. LLMs hit rate limits. We handle all of these gracefully with detailed logging and user-friendly messages.
[PAUSE:2500]
-->

# Error Handling & Resilience 🛡️

**Graceful Degradation** —

```go
// If source unreachable, mark failed
if err := fetchURL(url); err != nil {
    return VerificationResult{
        Verified: false,
        Reason:   "Source unreachable",
    }
}
```

**Retry Logic** —
- HTTP retries with exponential backoff
- Automatic quality check retries
- Human-in-the-loop for partial results

**User-Friendly Messages** —
- "Found 8 of 10 requested, continue? (y/n)"
- Clear error messages with remediation steps

---

<!--
Security considerations went beyond API keys. We implemented request timeouts to prevent resource exhaustion. Input validation prevents injection attacks. Rate limiting could be added at the reverse proxy level. And all secrets are environment-based, never in code.
[PAUSE:2500]
-->

# Security Considerations 🔐

**API Key Management** —
- Environment variables only (never in code)
- Server-side storage (clients don't need keys)
- Per-agent key rotation possible

**Input Validation** —
- Topic length limits (500 chars)
- Min/max stats bounds (1-100)
- URL validation before fetching

**Timeouts** —
- HTTP request timeouts (30-60s)
- LLM generation timeouts
- Overall query timeout (120s)

**Future**: Add rate limiting, authentication

---

<!--
Performance optimization was iterative. We profiled each agent. Added caching where appropriate. Optimized LLM prompts to reduce tokens. And parallelized independent operations. There's always room for improvement, but we've achieved good performance.
[PAUSE:2500]
-->

# Performance Optimization 🚄

**Research Agent** —
- Parallel URL searches where supported
- Connection pooling for HTTP clients

**Synthesis Agent** —
- Parallel page fetching (up to 5 concurrent)
- Content truncation (30K chars max)
- Efficient JSON parsing

**Verification Agent** —
- Batch verification where possible
- Early exit on clear failures
- LLM only for fuzzy matching

**Overall** —
- 45-second average for 10 verified statistics
- Scales linearly with min_stats target

---

<!--
The code structure promotes maintainability. Shared models prevent drift. The package organization is clear. Agent independence means you can refactor one without breaking others. And the factory patterns make adding new providers trivial.
[PAUSE:2500]
-->

# Code Organization 📁

```
agents/          # Each agent is independent
  ├── research/
  ├── synthesis/
  ├── verification/
  ├── direct/
  └── orchestration-eino/

pkg/             # Shared libraries
  ├── config/    # Centralized configuration
  ├── llm/       # Multi-provider factory
  ├── models/    # Shared data structures
  ├── search/    # Search abstraction
  └── direct/    # Direct search service

main.go          # CLI entry point
```

**Principle**: High cohesion, low coupling

---

<!--
Documentation was a first-class citizen. The README is comprehensive with clear warnings. Each agent has inline comments. OpenAPI docs for the Direct agent. And this presentation serves as architectural documentation. Good docs enable adoption.
[PAUSE:2500]
-->

# Documentation Strategy 📚

**README.md** —
- Comprehensive setup instructions
- Clear mode comparisons
- Warning callouts for limitations

**Code Documentation** —
- Inline comments for complex logic
- Function documentation (godoc format)
- Architecture decision records (ADRs)

**API Documentation** —
- OpenAPI 3.1 specification (Huma)
- Interactive Swagger UI at `/docs`
- Example requests/responses

**Presentation**: Architecture overview (this!)

---

<!--
Community and extensibility were design goals. The multi-provider support means teams can use their preferred LLM. The modular architecture means you can swap out agents. And the open-source license encourages contributions.
[PAUSE:2500]
-->

# Extensibility & Contributions 🤝

**Easy to Extend** —
- Add new LLM provider: Implement `gollm` interface
- Add new search provider: Implement `metasearch` interface
- Add new agent: Follow existing patterns
- Add new verification rules: Extend verification agent

**Contribution Areas** —
- 🔧 New LLM providers (e.g., Perplexity)
- 🌍 Multi-language support
- 📊 Range statistics (`value_max`)
- ⚡ Performance optimizations
- 📚 Documentation improvements

**License**: MIT (permissive)

---

<!--
Real-world usage patterns emerged. The pipeline mode is used for research reports and data analysis. Direct mode is used for quick brainstorming. The MCP integration is used by AI assistants. Different modes serve different needs.
[PAUSE:2500]
-->

# Real-World Usage Patterns 🌐

1. **Use Case 1**: Research Reports
    - Pipeline mode with `--reputable-only`
    - Export to JSON for analysis
    - Cite sources with URLs
2. **Use Case 2**: Data Analysis
    - Bulk queries via API
    - Process results in pandas/R
    - Visualization of trends
3. **Use Case 3**: AI Assistant Integration
    - MCP server with Claude Code
    - LLM asks stats-agent for verified data
    - Compose into reports
4. **Use Case 4**: Quick Fact-Checking
    - Direct mode for fast lookup
    - Accept unverified for speed

---

<!--
The cost model is important for production. L L M A P I costs dominate. Search A P I costs are secondary. But the value is in accuracy. One wrong statistic in a report can be costly. We provide cost performance tradeoffs.
[PAUSE:1500]
Let's be honest about costs. Direct mode costs about one cent per query because it's a single L L M call. Hybrid mode, which uses the L L M to generate statistics then verifies them, costs about three cents. Pipeline mode, the full system with search, synthesis, and verification, costs around ten cents per query. The breakdown: search A P I costs two cents for thirty U R Ls. L L M calls for extracting statistics from fifteen pages and verifying them cost about eight cents total. The main cost driver is how many pages you process. Using Gemini two point five Flash instead of G P T four or Claude reduces costs significantly because Gemini is cheaper per token. But here's the key question: what's the cost of using wrong statistics in your research report or business presentation? Ten cents for verified accuracy is cheap insurance.
[PAUSE:2500]
-->

# Cost Analysis 💰

**Per Query Costs** (estimates):

| Component | Direct | Hybrid | Pipeline |
|-----------|--------|--------|----------|
| Search API | $0.00 | $0.00 | $0.02 |
| LLM Calls | $0.01 | $0.03 | $0.08 |
| **Total** | **$0.01** | **$0.03** | **$0.10** |

**Cost Drivers** —
- Number of pages processed (15+)
- LLM provider choice (Gemini < Claude < GPT-4o/GPT-5)
- Verification attempts

**Optimization**: Use Gemini 2.5 Flash (fast + cheap)

---

<!--
Scaling considerations matter for high-volume use. Each agent can be independently scaled. Add load balancers in front. Use a message queue for async processing. Cache search results. These patterns enable production deployment.
[PAUSE:1500]
The modular architecture really shines when you need to scale. Each agent type runs independently, so you can scale them horizontally based on load. If synthesis is the bottleneck because L L M calls are slow, run twenty synthesis agents behind a load balancer while keeping ten orchestrators. If you need higher throughput, scale vertically by increasing concurrency limits, processing larger content chunks, or fetching more pages in parallel. For even better performance at scale, add a caching layer for search results with a one hour T T L, so repeated queries for the same topic use cached U R Ls. Use a message queue like Rabbit M Q or Kafka for async bulk processing. Store results in a database for analytics. The architecture supports all these patterns because the agents are stateless and communicate via H T T P. Going from ten queries per minute to thousands just requires infrastructure, not code changes.
[PAUSE:2500]
-->

# Scaling Considerations 📈

**Horizontal Scaling** —
- Each agent scales independently
- Load balancer per agent type
- Stateless design enables easy scaling

**Vertical Scaling** —
- Increase concurrency limits
- Larger content chunks (current: 30K)
- More parallel page fetching

**Optimizations for Scale** —
- Cache search results (1 hour TTL)
- Queue-based processing for bulk queries
- Database for results persistence

**Example**: 10 orchestrators + 20 synthesis agents

---

<!--
Monitoring in production needs more than logs. We'd add metrics collection. Track verification rates over time. Alert on degraded performance. Distributed tracing would help debug issues. These are standard production practices.
[PAUSE:2500]
-->

# Production Monitoring 📡

1. **Metrics to Collect**
    - Verification rate by source domain
    - Response time percentiles (p50, p95, p99)
    - Error rate by agent
    - API cost per query
    - Throughput (queries/minute)
2. **Alerting**
    - Verification rate < 50% (alert)
    - Response time > 120s (alert)
    - Agent health check failures
    - API quota exhaustion
3. **Tools** (future)
    - Prometheus for metrics
    - Grafana for dashboards
    - Jaeger for distributed tracing

---

<!--
Compliance matters for some use cases. We cite sources properly. Respect robots.txt. Rate limit our fetching. Store only necessary data. These practices ensure we're a good web citizen and legally compliant.
[PAUSE:2500]
-->

# Compliance & Ethics 🌟

**Responsible Web Scraping** —
- Respect `robots.txt`
- Rate limiting on URL fetches
- User-Agent identification
- No aggressive crawling

**Data Privacy** —
- No PII collection
- No user query logging (optional)
- API keys stored securely
- GDPR compliance considerations

**Source Attribution** —
- Always cite original sources
- Provide full URLs
- Verbatim excerpts (fair use)

**Ethics**: Promote verified information, combat misinformation

---

<!--
Let's talk about the competitive landscape. We compared our approach to several alternatives. Each has tradeoffs. Our system uniquely combines real-time search with rigorous verification in an open architecture.
[PAUSE:1500]
How do we stack up against existing solutions? Chat G P T dot com does search using Bing and has light verification, but only supports G P T models and is closed source. Perplexity uses multiple search providers with light verification, but has limited L L M options and is also closed. Direct L L M usage, just asking an L L M without search, has no verification at all, though you can use any L L M. Our system stands out in two ways. First, we have strong verification, actually fetching sources and validating excerpts programmatically. Second, we're completely open source under M I T license and support five plus L L M providers. The community can audit our code, see exactly how verification works, extend it for their needs, and choose their preferred L L M and search providers. Transparency and flexibility are our competitive advantages.
[PAUSE:2500]
-->

# Competitive Analysis 🏆

| System | Search | Verify | Multi-LLM | Open Source |
|--------|--------|--------|-----------|-------------|
| **ChatGPT.com** | ✅ Bing | ⚠️ Light | ❌ GPT only | ❌ Closed |
| **Perplexity** | ✅ Multiple | ⚠️ Light | ❌ Limited | ❌ Closed |
| **Our System** | ✅ Google | ✅ **Strong** | ✅ 5+ | ✅ **MIT** |
| **Direct LLM** | ❌ Memory | ❌ None | ✅ Any | N/A |

**Key Differentiator**: Rigorous verification + flexibility

**Open Source**: Community can audit, extend, trust

---

<!--
Migration from existing systems is straightforward. If you're using direct LLM calls, switch to our Direct agent for server-side security. If you're using ChatGPT API, use our Pipeline mode for verification. The API is simple and RESTful.
[PAUSE:2500]
-->

# Migration Path 🚀

**From Direct LLM Usage** —
```python
# Before: Client-side LLM
response = openai.chat("Find climate statistics")

# After: Stats Agent Direct mode
response = requests.post(
    "http://localhost:8005/search",
    json={"topic": "climate change", "min_stats": 10}
)
```

**From ChatGPT API** —
```python
# Before: ChatGPT (no verification)
stats = ask_chatgpt("climate statistics")

# After: Stats Agent Pipeline (verified)
response = requests.post(
    "http://localhost:8000/orchestrate",
    json={"topic": "climate change", "min_verified_stats": 10}
)
```

---

<!--
The roadmap ahead includes several exciting features. Perplexity integration would give us built-in search. Streaming responses would improve perceived performance. Range statistics would handle more data types. And multi-language support would expand our reach.
[PAUSE:1500]
Looking ahead, we have an exciting roadmap. Q one twenty twenty-five priorities include integrating Perplexity A P I, which has built-in search so we wouldn't need separate search providers, adding a value max field to support range statistics like "seventy nine to ninety six percent," and implementing response streaming so users see results as they're found rather than waiting for everything. Q two focuses on multi-language support for Spanish, French, German, and Chinese sources, a caching layer to reduce redundant searches and costs, and a Graph Q L A P I option for more flexible querying. Q three gets ambitious with a browser extension for real-time fact checking as you browse, integrations with Notion and Confluence for embedding verified statistics in documentation, and advanced citation formats like A P A and M L A for academic use. This roadmap is community driven. Submit feature requests on GitHub, and we'll prioritize based on demand.
[PAUSE:2500]
-->

# Roadmap 🗺️

**Q1 2025** —
- ✨ Perplexity API integration (built-in search)
- ✨ Range statistics (`value_max` field)
- ✨ Response streaming for faster UX

**Q2 2025** —
- ✨ Multi-language support (ES, FR, DE, ZH)
- ✨ Caching layer for search results
- ✨ GraphQL API option

**Q3 2025** —
- ✨ Browser extension for fact-checking
- ✨ Notion/Confluence integrations
- ✨ Advanced citation formats (APA, MLA)

**Community Driven**: Submit feature requests on GitHub!

---

<!--
Team collaboration was key to success. Clear architecture boundaries meant parallel development. Regular sync meetings kept us aligned. Code reviews maintained quality. And documentation ensured knowledge transfer.
[PAUSE:2500]
-->

# Team & Collaboration 👥

**Development Approach** —
- Agent-based architecture enables parallel work
- Clear interfaces between components
- Code reviews for quality
- Continuous integration (GitHub Actions)

**Best Practices** —
- Branch protection on main
- Required passing tests for merge
- Semantic versioning
- Changelog maintenance

**Communication** —
- Architecture decisions documented
- Weekly sync meetings
- GitHub issues for tracking

---

<!--
Lessons learned extend beyond code. Start with clear requirements. Build verification early, not as an afterthought. Be honest about limitations. And always prioritize user experience. These principles apply to any multi-agent system.
[PAUSE:1500]
These eleven lessons fall into three categories. Technical lessons: real-time data beats L L M memory, verification is essential not optional, modular architecture enables optimization, and prompt engineering is critical at scale. Process lessons: clear requirements prevent scope creep, early testing reveals issues sooner, documentation enables adoption, and user feedback drives priorities. Product lessons: be honest about limitations to build trust, provide flexibility to drive adoption through multi L L M and multi search support, and prioritize developer experience because if it's hard to use locally, it won't get used in production. These aren't just lessons for statistics agents. They apply to any A I system, any multi-agent architecture, any production service. Architecture, process, and product thinking matter as much as code quality.
[PAUSE:2500]
-->

# Lessons Learned (Summary) 💭

1. **Technical**
    <ol type="1">
    <li>Real-time data > LLM memory for facts</li>
    <li>Verification is essential, not optional</li>
    <li>Modular architecture enables optimization</li>
    <li>Prompt engineering is critical at scale</li>
    </ol>
2. **Process**
    <ol type="1" start="5">
    <li>Clear requirements prevent scope creep</li>
    <li>Early testing reveals issues sooner</li>
    <li>Documentation enables adoption</li>
    <li>User feedback drives priorities</li>
    </ol>
3. **Product**
    <ol type="1" start="9">
    <li>Be honest about limitations (builds trust)</li>
    <li>Provide flexibility (multi-LLM, multi-search)</li>
    <li>Developer experience matters</li>
    </ol>

---

<!--
Closing thoughts: Building a multi-agent system is challenging but rewarding. The key is clear separation of concerns. Each agent does one thing well. Together they create something greater than the sum of parts. This architecture pattern applies to many domains.
[PAUSE:1500]
Let's wrap up with what we've built and what it means. We created a production-ready system that achieves sixty to ninety percent verification of statistics, compared to zero percent for L L Ms answering from memory. That's the fundamental value proposition: trust. Researchers, journalists, and analysts can now use A I to find statistics and actually trust the results because we provide verifiable sources. The multi-agent architecture with clear separation, research, synthesis, verification, orchestration, enables independent optimization and debugging. The flexibility to use any L L M provider or search provider means different organizations can adopt based on their constraints. And being open source under M I T license means the community can audit, extend, and trust the system. This project proves that with the right architecture, you can combine the intelligence of L L Ms with the accuracy of real-time verification.
[PAUSE:2500]
-->

# Conclusion 🎓

**What We Built** —
- Production-ready statistics verification system
- 60-90% verification rate (vs 0% for LLM alone)
- Multi-agent architecture with clear separation
- Flexible (multi-LLM, multi-search)
- Open source (MIT license)

**Key Success Factors** —
- Real-time web search for current data
- Rigorous verification against sources
- Modular, extensible design
- Comprehensive testing & documentation

**Impact**: Enables verified statistics for research, reporting, analysis

---

<!--
We welcome contributions from the community. Whether it's adding a new LLM provider, fixing a bug, improving documentation, or suggesting features - all contributions are valuable. Check out our GitHub repository to get started.
[PAUSE:2500]
-->

# Get Involved! 🚀

**Repository**: `github.com/grokify/stats-agent-team`

**Quick Start** —
```bash
git clone https://github.com/grokify/stats-agent-team
cd stats-agent-team
make install
make build
make run-all-eino
```

**Contribute** —
- 🐛 Report bugs
- 💡 Suggest features
- 📝 Improve docs
- 🔧 Submit PRs

**License**: MIT (permissive, commercial-friendly)

---

<!--
Thank you for your attention. We've covered the journey from requirements to a working system. The challenges we faced, the solutions we implemented, and the lessons we learned. We hope this inspires your own multi-agent projects. Questions?
[PAUSE:1500]
We've covered a lot today. From the fundamental problem of L L M hallucinations to the architecture that solves it. From zero percent verification in direct mode to sixty to ninety percent in pipeline mode. From single L L M support to five plus providers. From client-side insecurity to proper server-side architecture. From J S O N parsing bugs to comprehensive prompt engineering. Every challenge taught us something. Every solution opened new possibilities. If you want to try it yourself, the repo is on GitHub at github dot com slash grokify slash stats dash agent dash team. The documentation is comprehensive, setup is straightforward, and we welcome contributions. Special thanks to the Google A D K team, Eino framework contributors, and the entire open source community. Now, let's open it up for questions.
[PAUSE:2000]
-->

# Questions? 🤔

**Contact & Resources** —
- 📧 GitHub Issues for questions
- 📚 Full documentation in README.md
- 🔗 OpenAPI docs at `localhost:8005/docs`
- 💬 Discussions tab for community chat

**Thank You!** 🙏

**Special Thanks** —
- Google ADK team
- Eino framework contributors
- Open source LLM providers
- The Go community

---

<!--
For those interested in diving deeper, we have comprehensive documentation. The README covers setup and usage. The architecture document explains design decisions. And the API documentation provides integration details. All available in the repository.
[PAUSE:2000]
-->

# Additional Resources 📖

**Documentation** —
- `README.md` - Setup & usage guide
- `4_AGENT_ARCHITECTURE.md` - Architecture deep dive
- `LLM_CONFIGURATION.md` - Multi-LLM setup
- `SEARCH_INTEGRATION.md` - Search provider setup
- `MCP_SERVER.md` - MCP integration guide
- `DOCKER.md` - Container deployment

**Example Queries** —
- Climate change statistics
- AI industry trends
- Healthcare outcomes
- Economic indicators
- Educational metrics

**Try it yourself!** 🚀
