# hevy-cli

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![CLI](https://img.shields.io/badge/interface-terminal-111827)](#quickstart)
[![Hevy](https://img.shields.io/badge/data-Hevy-ef4444)](https://www.hevyapp.com)
[![WHOOP](https://img.shields.io/badge/recovery-WHOOP-22c55e)](#readiness--recovery)

**A fast, scriptable, AI-friendly terminal client for Hevy.**

Track workouts, inspect routines, export data, analyze progress, detect plateaus, and combine training history with WHOOP recovery — all from the command line.

</div>

---

## Why this exists

Hevy has a great mobile app, but sometimes you want terminal-native access:

- check your last workout without opening your phone
- export your data for scripting or analysis
- search exercises and routines quickly
- inspect weekly volume, PRs, and training split
- plug your training data into **OpenClaw**, **Claude Code**, **Codex**, or your own agent workflows

`hevy-cli` gives you that layer.

---

## Highlights

### Core Hevy access
- workouts, workout counts, latest workout, today’s workout
- routines and routine details
- exercise template search + exercise details
- machine-readable export (`csv`, JSON output)

### Training analytics
- streak tracking
- PR detection
- weekly summary
- workout diffing
- volume charts
- muscle hit map
- workout calendar
- workout search
- progression charts

### Advanced insights
- next workout suggestion (`plan`)
- training consistency reports
- plateau detection
- superset usage analysis
- fatigue signal detection from RPE
- actual split analysis
- all-time records vs current bests
- time-efficiency / rest proxy analysis

### Recovery integration
- WHOOP-backed readiness scoring
- recovery + training recommendations
- recent recovery trend display

### Agent-friendly UX
- clean terminal output
- JSON output for automation
- compact output mode
- shell completions
- works well in scripts, TUI workflows, and coding agents

---

## Quickstart

### Install

#### Go install
```bash
go install github.com/dhruvkelawala/hevy-cli@latest
```

#### Build from source
```bash
git clone https://github.com/dhruvkelawala/hevy-cli.git
cd hevy-cli
go build -o hevy .
```

---

## Setup

### Interactive setup
```bash
hevy init
```

### Or use an environment variable
```bash
export GO_HEVY_API_KEY="your-hevy-api-key"
```

### Or store it in config
```bash
hevy config set api_key "your-hevy-api-key"
```

Config precedence:
1. `GO_HEVY_API_KEY`
2. local config file

Default config path:
```bash
~/Library/Application Support/hevy-cli/config.json
```

---

## First 60 seconds

```bash
# verify auth
hevy status

# who am i?
hevy me

# recent workouts
hevy workouts --page-size 5

# latest workout in detail
hevy last

# this week’s summary
hevy week

# personal records
hevy pr --all
```

---

## Command map

## Core commands

| Command | What it does |
|---|---|
| `hevy status` | Verify API access and show account summary |
| `hevy me` | Show authenticated user profile |
| `hevy count` | Show total workout count |
| `hevy workouts` | List recent workouts |
| `hevy workout <id>` | Show or mutate a workout |
| `hevy last` | Show the most recent workout in detail |
| `hevy today` | Show today’s workout if one exists |
| `hevy routines` | List routines |
| `hevy routine <id>` | Show routine details |
| `hevy exercises` | List/search exercise templates |
| `hevy exercise <id>` | Show exercise template details |
| `hevy history <exercise-id>` | Show exercise history |
| `hevy export --format csv` | Export workouts in machine-readable format |

### Insight commands

| Command | What it does |
|---|---|
| `hevy progress "Squat"` | ASCII progression chart for an exercise |
| `hevy streak` | Weekly workout streak tracker |
| `hevy pr` | Personal records by exercise |
| `hevy week` | Weekly training summary |
| `hevy diff` | Compare two workouts (defaults to last two) |
| `hevy volume "Squat"` | Volume-over-time chart |
| `hevy muscles` | Muscle groups hit this week |
| `hevy calendar` | ASCII workout calendar |
| `hevy search upper` | Search workouts by title or exercise name |

### Advanced analytics

| Command | What it does |
|---|---|
| `hevy plan` | Suggest your next workout |
| `hevy consistency` | Show training consistency over time |
| `hevy plateau` | Detect stalled exercises |
| `hevy supersets` | Show most-used superset pairings |
| `hevy fatigue` | Analyze RPE trends for fatigue signals |
| `hevy split` | Analyze your actual training split |
| `hevy records` | Compare all-time vs current bests |
| `hevy rest` | Estimate workout time-efficiency |
| `hevy readiness` | Combine recovery + training readiness |

### Utility commands

| Command | What it does |
|---|---|
| `hevy config` | Show/update configuration |
| `hevy completion zsh` | Generate shell completions |
| `hevy version` | Print version |
| `hevy init` | Interactive configuration |

---

## Practical examples

### Training overview
```bash
hevy last
hevy week
hevy streak
hevy calendar
```

### Strength tracking
```bash
hevy pr --all
hevy progress "Bench Press"
hevy volume "Squat"
hevy plateau
hevy records
```

### Programming your next session
```bash
hevy muscles
hevy split
hevy plan
hevy readiness
```

### Data extraction
```bash
hevy workouts --json
hevy exercises --search curl --json
hevy export --format csv > workouts.csv
```

### Search and compare
```bash
hevy search upper
hevy diff
hevy history <exercise-id>
```

---

## Readiness & recovery

`hevy-cli` can use WHOOP data to make training recommendations.

### Configure WHOOP integration
```bash
hevy config set whoop_path /path/to/whoop-tracker
```

### Run readiness
```bash
hevy readiness
```

Example output:
```text
🟢 WHOOP Recovery: 89% | HRV: 78ms | RHR: 56bpm
Status: GREEN — full send. heavy compounds ok.
Suggested: Pull
```

This is especially useful if you want your recovery state and recent training history in one place.

---

## AI Agents

`hevy-cli` is built to work well with agents because it is:
- scriptable
- terminal-native
- JSON-capable
- deterministic enough for repeated command execution

### OpenClaw
Great fit for:
- workout summaries
- routine inspection
- trend analysis
- daily readiness checks
- automated logging / reminders

Examples:
```bash
hevy week --json
hevy readiness
hevy pr --all --json
```

### Claude Code / Codex / Cursor / other coding agents
Useful when you want an agent to:
- inspect your recent training state
- generate charts or reports from raw JSON
- compare workouts
- identify plateaus or consistency issues
- build custom tooling on top of Hevy data

Examples:
```bash
hevy workouts --json > workouts.json
hevy exercises --json > exercises.json
hevy records --json
hevy consistency --json
```

### Recommended agent workflow
```bash
# 1. fetch data
hevy workouts --json > workouts.json
hevy week --json > week.json
hevy readiness --json > readiness.json

# 2. let your agent analyze/report from those files
```

---

## Output modes

Most commands support:
- default table/text output
- `--json` / `-j`
- `--compact`
- `--kg`
- `--lbs`

Examples:
```bash
hevy workouts --json
hevy last --compact
hevy volume "Squat" --lbs
```

---

## Pagination

List commands support paging:

```bash
hevy workouts --page 2 --page-size 5
hevy routines --page-size 10
hevy exercises --page-size 25
hevy workouts --limit 20
hevy workouts --all
```

Current API limits:
- workouts: max 10
- routines: max 10
- exercises: max 100

---

## Create / update workout JSON shape

```json
{
  "workout": {
    "title": "Friday Leg Day 🔥",
    "description": "Medium intensity leg day focusing on quads.",
    "start_time": "2024-08-14T12:00:00Z",
    "end_time": "2024-08-14T12:45:00Z",
    "is_private": false,
    "exercises": [
      {
        "exercise_template_id": "D04AC939",
        "notes": "Felt strong today.",
        "sets": [
          {
            "type": "normal",
            "weight_kg": 100,
            "reps": 10,
            "rpe": 8.5
          }
        ]
      }
    ]
  }
}
```

---

## Shell completions

```bash
hevy completion zsh > ~/.zsh/completions/_hevy
hevy completion bash > /etc/bash_completion.d/hevy
hevy completion fish > ~/.config/fish/completions/hevy.fish
```

---

## Development

```bash
go build ./...
go test ./...
go vet ./...
```

---

## Roadmap ideas

- richer charts and sparklines
- better exercise-history UX around IDs
- richer export formats
- release binaries + package manager distribution polish
- more AI-agent example workflows

---

## License

MIT
