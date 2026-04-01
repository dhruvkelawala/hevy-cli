# hevy-cli

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/dhruvkelawala/hevy-cli?color=22c55e)](https://github.com/dhruvkelawala/hevy-cli/releases)
[![Hevy](https://img.shields.io/badge/data-Hevy-ef4444)](https://www.hevyapp.com)
[![WHOOP](https://img.shields.io/badge/recovery-WHOOP-22c55e)](#-whoop-readiness)

<img src="./assets/hevy-cli-hero.png" alt="hevy-cli — crab lifting weights" width="100%" />

**Give your AI agent a gym buddy.**

A terminal client for [Hevy](https://www.hevyapp.com) built for AI coding agents — deterministic `--json` output, single-binary install, and a [SKILL.md](./SKILL.md) that any agent can pick up instantly.

<p>
  <a href="#-for-agents"><strong>Agent Setup</strong></a> ·
  <a href="#-what-can-it-do"><strong>Features</strong></a> ·
  <a href="#-whoop-readiness"><strong>WHOOP</strong></a>
</p>

</div>

---

## 🚀 Install

```bash
# macOS / Linux — download latest binary
curl -sL https://github.com/dhruvkelawala/hevy-cli/releases/latest/download/hevy_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz | tar xz -C /tmp && mv /tmp/hevy ~/.local/bin/

# Or via Go
go install github.com/dhruvkelawala/hevy-cli@latest
```

Then authenticate:

```bash
hevy init
# or: export GO_HEVY_API_KEY="your-key"
```

---

## 🤔 What can it do?

### See your training at a glance

```
$ hevy week
sessions          2
total_volume      19,599kg
avg_duration_min  73.4
muscle_groups     chest(1), lats(1), biceps(1), triceps(1), quads(1), hamstrings(1)

$ hevy streak
Current streak: 5 weeks (since Mar 2)

$ hevy calendar
March 2026
Mo Tu We Th Fr Sa Su
[2]  3  4  5  6 [7]  8
[9] [10] 11 [12] 13 14 15
[16] 17 [18] [19] 20 21 22
23 24 [25] 26 [27] 28 29
[30] [31]
```

### Track your strength

```
$ hevy pr --all
Squat (Barbell)            70.00kg × 12
Barbell Bench Press        55.00kg × 8
Bent Over Row              50.00kg × 12
Chest Press (Machine)      40.00kg × 12

$ hevy progress "Squat"
Squat (Barbell) — last 2 sessions
Mar 18  65.00kg  █████████████
Mar 30  70.00kg  ██████████████

$ hevy volume "Squat"
Squat (Barbell) volume
Mar 18  2,080kg  ████████
Mar 30  3,420kg  ██████████████
```

### Know what to train next

```
$ hevy plan
Last trained:
  chest       0 days ago (Tue)
  back        0 days ago (Tue)
  legs        1 day ago (Mon)
  shoulders   5 days ago (Thu)

Suggested next: Pull
Reason: shoulders haven't been hit in 5 days

$ hevy split
TYPE   COUNT  PERCENTAGE
Legs       3  27%
Upper      3  27%
Pull       2  18%
Push       1   9%
```

### Detect problems before they stall you

```
$ hevy plateau
Bench Press (Barbell)     50.00kg  stalled 4 sessions
Overhead Press            35.00kg  stalled 3 sessions

$ hevy consistency
Mar 2026   12 sessions   75% hit rate
Feb 2026    0 sessions    0%

$ hevy fatigue
No RPE data found. Log RPE in Hevy to enable fatigue tracking.
```

### 🟢 WHOOP Readiness

Combine your WHOOP recovery data with training history for a single "should I train today?" answer.

```
$ hevy readiness
🟢 WHOOP Recovery: 89%  |  HRV: 78ms  |  RHR: 56bpm

Status: GREEN — full send. heavy compounds ok.
Last trained: Upper (Mar 31)
Avoid today: back, biceps, chest, shoulders, triceps
Suggested: Pull

Recovery history (last 7 days):
Mon  89%  ████████
Sun  39%  ███
Sat  99%  █████████
Fri  66%  ██████
```

```bash
# Setup: point to your WHOOP tracker skill
hevy config set whoop_path /path/to/whoop-tracker
```

---

## 🤖 For Agents

hevy-cli exists because AI agents need structured fitness data, not screenshots of an app. Every command supports `--json`, output is deterministic, and installation is a single `curl`.

### Why agents need this

- **Morning briefings** — your agent pulls WHOOP recovery + last workout + weekly stats and tells you whether to train today
- **Plateau detection** — agent spots stalled lifts before you do, suggests deload or variation
- **Accountability** — agent checks `hevy streak` and nudges you if you're slipping
- **Programming** — agent reads your split, volume trends, and recovery to suggest what to train next

### Agent quickstart

```bash
# 1. Drop SKILL.md into your agent's workspace
cp SKILL.md ~/.openclaw/workspace/skills/hevy-cli/

# 2. Set the API key
export GO_HEVY_API_KEY="your-hevy-api-key"

# 3. Agent can now use any command
hevy readiness --json    # Should I train today?
hevy plan --json         # What should I train?
hevy plateau --json      # Am I stalling anywhere?
hevy week --json         # Weekly summary
```

### Works with

- **[OpenClaw](https://openclaw.ai)** — drop SKILL.md into skills directory, agent auto-discovers it
- **[Claude Code](https://docs.anthropic.com/en/docs/claude-code)** / **[Codex](https://openai.com/index/codex/)** — pass `--json` flag, pipe output into agent context
- **Any agent framework** — it's just a binary with structured output, no SDK needed

---

## 📋 All 30+ Commands

<details>
<summary><strong>Core</strong> — workouts, exercises, routines</summary>

| Command | What it does |
|---|---|
| `hevy status` | Verify API access |
| `hevy me` | User profile |
| `hevy count` | Total workout count |
| `hevy workouts` | List recent workouts |
| `hevy workout <id>` | Show/mutate a workout |
| `hevy last` | Most recent workout |
| `hevy today` | Today's workout |
| `hevy routines` | List routines |
| `hevy routine <id>` | Routine details |
| `hevy exercises` | Search exercise templates |
| `hevy exercise <id>` | Exercise details |
| `hevy history <id>` | Exercise history |
| `hevy export --format csv` | Export all workouts |

</details>

<details>
<summary><strong>Analytics</strong> — track, compare, visualize</summary>

| Command | What it does |
|---|---|
| `hevy progress "Squat"` | ASCII progression chart |
| `hevy streak` | Weekly streak tracker |
| `hevy pr --all` | All personal records |
| `hevy week` | Weekly summary |
| `hevy diff` | Compare last two workouts |
| `hevy volume "Squat"` | Volume over time |
| `hevy muscles` | Muscle groups hit this week |
| `hevy calendar` | ASCII workout calendar |
| `hevy search upper` | Search workouts |

</details>

<details>
<summary><strong>Advanced</strong> — plan, detect, recover</summary>

| Command | What it does |
|---|---|
| `hevy plan` | Suggest next workout |
| `hevy consistency` | Training consistency report |
| `hevy plateau` | Detect stalled exercises |
| `hevy supersets` | Superset pairings |
| `hevy fatigue` | RPE trend analysis |
| `hevy split` | Actual training split |
| `hevy records` | All-time vs current bests |
| `hevy rest` | Time efficiency |
| `hevy readiness` | WHOOP + training readiness |

</details>

<details>
<summary><strong>Utility</strong></summary>

| Command | What it does |
|---|---|
| `hevy config` | Show/update config |
| `hevy config set` | Set a config value |
| `hevy completion zsh` | Shell completions |
| `hevy init` | Interactive setup |
| `hevy version` | Print version |

</details>

---

## ⚙️ Output modes

```bash
hevy workouts              # table (default)
hevy workouts --json       # JSON
hevy workouts --compact    # one-liner
hevy last --lbs            # pounds
hevy last --kg             # kilograms
```

---


## 🛠️ Development

```bash
git clone https://github.com/dhruvkelawala/hevy-cli.git
cd hevy-cli
go build ./...
go test ./...
go vet ./...
```

Releases are built automatically via [GoReleaser](https://goreleaser.com/) on tag push.

---

## License

MIT
