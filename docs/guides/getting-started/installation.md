# Installation

## Prerequisites

vac requires the following tools to be installed:

### 1. Go 1.21+

```bash
go version
# Should output: go version go1.21.x or higher
```

### 2. ffmpeg

Used for video recording and processing.

=== "macOS"

    ```bash
    brew install ffmpeg
    ```

=== "Linux (Debian/Ubuntu)"

    ```bash
    sudo apt install ffmpeg
    ```

=== "Windows"

    Download from [ffmpeg.org](https://ffmpeg.org/download.html) and add to PATH.

### 3. Marp CLI

Used for rendering Markdown to HTML presentations.

```bash
npm install -g @marp-team/marp-cli
```

### 4. ElevenLabs API Key

1. Sign up at [ElevenLabs](https://elevenlabs.io/)
2. Get your API key from the dashboard
3. Set the environment variable:

```bash
export ELEVENLABS_API_KEY="your-api-key-here"
```

## Install vac

### From Source

```bash
git clone https://github.com/grokify/videoascode
cd vac
go build -o bin/vac ./cmd/vac
```

### Using Go Install

```bash
go install github.com/grokify/videoascode/cmd/vac@latest
```

## Verify Installation

Check that all dependencies are available:

```bash
vac --check
```

This will verify:

- [x] ffmpeg is installed
- [x] Marp CLI is installed
- [x] ElevenLabs API key is set
