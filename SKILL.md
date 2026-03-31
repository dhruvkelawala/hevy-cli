---
name: hevy-cli
description: "Hevy workout tracking CLI. List workouts, exercises, routines, track progression, export data. Use when user asks about workouts, gym sessions, exercise history, fitness progress, or Hevy data."
---

# hevy-cli — Hevy Workout Tracker CLI

Binary: `hevy` (installed at `~/.local/bin/hevy`)

## Prerequisites
- `GO_HEVY_API_KEY` env var set (stored in `~/.zshrc`)
- For agent sessions: `source ~/.config/openclaw/env.sh`

## Quick Reference

### Check status
```bash
hevy status
```

### Last workout
```bash
hevy last
```

### List workouts
```bash
hevy workouts              # Recent 5
hevy workouts --limit 10   # Last 10
hevy workouts --all        # Everything
hevy workouts --json       # JSON output
```

### Workout detail
```bash
hevy workout <id>          # Full detail with exercises + sets
hevy workout <id> --json
```

### Exercises
```bash
hevy exercises --search "bench press"   # Search by name
hevy exercises --muscle chest           # Filter by muscle group
hevy exercises --custom                 # Custom exercises only
```

### Exercise history
```bash
hevy history <exercise-template-id>
hevy history D04AC939                   # Squat (Barbell) history
```

### Progression chart
```bash
hevy progress "Bench Press"             # ASCII chart of weight over time
hevy progress "Squat (Barbell)"
```

### Routines
```bash
hevy routines
hevy routine <id>
```

### Export
```bash
hevy export --format csv                # All workouts to CSV (stdout)
hevy export --format csv --output workouts.csv
```

### User info
```bash
hevy me
hevy count
```

## Common Agent Tasks

### "What did I do at the gym?"
```bash
hevy last
```

### "How's my bench press progressing?"
```bash
hevy progress "Bench Press"
```

### "Show my workouts this month"
```bash
hevy workouts --all
```

### "Find leg exercises"
```bash
hevy exercises --muscle quadriceps
hevy exercises --muscle hamstrings
```

## Output Flags
- Default: formatted table
- `--json` / `-j`: JSON for scripting
- `--compact`: one line per item
- `--kg` / `--lbs`: unit toggle (default kg)

## Notes
- API requires Hevy Pro subscription
- Exercise search paginates all pages (may take 1-2s)
- Progress does fuzzy name matching — use exact name for best results
- Repo: https://github.com/dhruvkelawala/hevy-cli
