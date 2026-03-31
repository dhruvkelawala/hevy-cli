# go-hevy

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

CLI tool for the [Hevy](https://www.hevyapp.com) workout tracking API, written in Go.

## Installation

```bash
go install github.com/dhruvkelawala/go-hevy@latest
```

Or download a binary from [Releases](https://github.com/dhruvkelawala/go-hevy/releases).

## Quick Start

```bash
# Set up your API key (get it from Hevy app settings)
hevy init

# List recent workouts
hevy workouts

# Get workout details
hevy workout <id>

# Check your profile
hevy me
```

## Commands

| Command | Description |
|---------|-------------|
| `hevy init` | Interactive setup — prompts for API key |
| `hevy me` | Show user profile |
| `hevy workouts` | List recent workouts |
| `hevy workout <id>` | Get workout details with exercises and sets |
| `hevy count` | Total workout count |
| `hevy routines` | List routines |
| `hevy routine <id>` | Get routine details |
| `hevy exercises` | List exercise templates |
| `hevy exercise <id>` | Get exercise template details |
| `hevy history <exercise-id>` | Exercise history with progression data |
| `hevy config` | Show current config |
| `hevy version` | Print version |

## Output Formats

```bash
# Default: human-readable table
hevy workouts

# JSON output for scripting
hevy workouts --json

# Compact one-line-per-item
hevy workouts --compact
```

## Configuration

Config is stored at `~/.config/go-hevy/config.json`.

You can also set the API key via environment variable:

```bash
export GO_HEVY_API_KEY=your-api-key-here
```

## API

go-hevy uses the [Hevy Public API](https://api.hevyapp.com/docs/). Get your API key from the Hevy app settings.

## Project Structure

```
go-hevy/
├── main.go
├── cmd/                    # CLI commands (cobra)
│   ├── root.go
│   ├── init.go
│   ├── workouts.go
│   ├── routines.go
│   ├── exercises.go
│   ├── history.go
│   ├── me.go
│   └── version.go
├── internal/
│   ├── api/                # HTTP client + API types
│   │   ├── client.go
│   │   ├── workouts.go
│   │   ├── routines.go
│   │   ├── exercises.go
│   │   └── types.go
│   ├── config/             # Config file management
│   │   └── config.go
│   └── output/             # Table + JSON formatters
│       ├── table.go
│       └── json.go
└── .github/workflows/      # CI + goreleaser
    └── release.yml
```

## Contributing

PRs welcome. Please open an issue first for major changes.

```bash
git clone https://github.com/dhruvkelawala/go-hevy.git
cd go-hevy
go build .
go test ./...
```

## License

MIT
