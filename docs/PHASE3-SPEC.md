# Phase 3: Advanced Analytics & WHOOP Integration

## Overview
Add 9 new commands to hevy-cli that provide deep training analytics and WHOOP recovery integration. All features must use only the existing Hevy API endpoints (no new API calls needed beyond what we already have) plus optional WHOOP data via a local Python script.

## Available Hevy API Endpoints (already implemented in our client)
- `GET /v1/workouts` (paginated, has all workout data including exercises, sets, weights, reps, timestamps)
- `GET /v1/workouts/{id}` (single workout detail)
- `GET /v1/workouts/count`
- `GET /v1/routines` (paginated)
- `GET /v1/routines/{id}`
- `GET /v1/exercise_templates` (paginated, includes primary_muscle_group, secondary_muscle_groups)
- `GET /v1/exercise_templates/{id}`
- `GET /v1/exercise_history/{id}` (set-level history for an exercise)

## Existing Codebase Structure
- `cmd/insights.go` — streak, muscles, week, diff, volume, search, calendar, pr commands
- `cmd/analytics.go` — weekSummary, workoutVolume, primaryMuscleForExercise, compareWorkouts helpers
- `cmd/helpers.go` — fetchWorkouts, fetchAllExercises, exerciseCatalogMap, pickExerciseByName, etc.
- `internal/api/` — API client, types
- `cmd/output/` — PrintTable, PrintKeyValueTable, PrintJSON helpers
- All commands follow the same pattern: cobra command → fetch data → compute → print (table/JSON)

## Important Conventions
- Every command must support `--json` flag for JSON output
- Every command must support `--kg` / `--lbs` unit flags
- Use `printLine()` for text output, `output.PrintTable()` for tables
- Use `fetchWorkouts(client, pageSize, limit, fetchAll)` — set fetchAll=true when you need history
- Use `exerciseCatalogMap(exercises)` to get muscle group info
- Use existing `parseAPITime()`, `workoutVolume()`, `workoutDurationMinutes()` helpers
- Register all new commands in `init()` at bottom of insights.go

## New Commands

### 1. `hevy plan`
**What**: Suggest next workout based on muscle recovery and training frequency.
**Logic**:
1. Fetch last 14 days of workouts
2. Build a map of muscle_group → last_trained_date using exercise catalog
3. Fetch user's routines
4. Score each routine by: how many days since those muscles were last hit (more days = higher score)
5. Recommend the top routine, with reasoning

**Output**:
```
Last trained:
  chest          1 day ago (Tue)
  back           1 day ago (Tue)  
  legs           2 days ago (Mon)
  shoulders      5 days ago (Thu)

Suggested next: Day 1 - Upper A (Push)
Reason: shoulders haven't been hit in 5 days, chest/triceps due for volume
```

**Flags**: `--days N` (lookback window, default 14)

### 2. `hevy consistency [--months N]`
**What**: Training consistency report over time.
**Logic**:
1. Fetch all workouts (or last N months)
2. Group by ISO week
3. Calculate: sessions per week, total volume per week, longest gap between sessions
4. Show monthly summary with target hit rate (assume 4 sessions/week target, configurable with `--target N`)

**Output**:
```
Monthly Consistency (last 3 months)
MONTH     SESSIONS  TARGET  HIT%   VOLUME      LONGEST GAP
Mar 2026        12    16    75%    58,240kg    3 days
Feb 2026        14    16    88%    62,100kg    2 days
Jan 2026        10    16    63%    45,800kg    5 days

Overall: 36/48 sessions (75%)
Current streak: 5 weeks without missing
```

**Flags**: `--months N` (default 3), `--target N` (sessions/week, default 4)

### 3. `hevy plateau [--threshold N]`
**What**: Detect stalled exercises where weight hasn't increased.
**Logic**:
1. Fetch all workouts
2. For each exercise that appears 3+ times, track max working weight per session
3. If weight hasn't increased in N consecutive sessions (default 3), flag as plateau
4. Sort by staleness (most sessions without improvement first)

**Output**:
```
EXERCISE                  WEIGHT   STALLED SINCE   SESSIONS
Bench Press (Barbell)     50.00kg  Mar 9           4 sessions
Overhead Press (Barbell)  35.00kg  Mar 9           3 sessions

2 exercises plateaued. Consider: deload, rep scheme change, or micro-loading.
```

**Flags**: `--threshold N` (min sessions to consider plateau, default 3), `--all` (show all exercises, not just plateaued)

### 4. `hevy supersets`
**What**: Show superset pairings from workout history.
**Logic**:
1. Fetch recent workouts (last 30 days)
2. Parse `superset_id` from exercises — exercises sharing the same non-nil superset_id in a workout are paired
3. Count frequency of each pairing
4. Display sorted by frequency

**Output**:
```
SUPERSET PAIR                                    COUNT  LAST USED
Chest Fly + Triceps Pushdown                         4  Mar 31
Lat Pulldown + Bicep Curl                            3  Mar 27
Face Pull + Lateral Raise                            2  Mar 19
```

**Note**: If no supersets found in data (user doesn't use them), print "No supersets found in recent workouts."

### 5. `hevy fatigue`
**What**: RPE trend analysis — detect accumulated fatigue.
**Logic**:
1. Fetch workouts with RPE data
2. For exercises where RPE is logged, track RPE vs weight over sessions
3. Flag exercises where RPE is trending up while weight is flat/down = fatigue signal
4. If no RPE data logged, print helpful message: "No RPE data found. Log RPE in Hevy to enable fatigue tracking."

**Output**:
```
EXERCISE                   RPE TREND    WEIGHT TREND   SIGNAL
Squat (Barbell)            8→8→9→9.5    70→70→70→70    ⚠️ FATIGUE
Bench Press                7→7→8        50→50→52       OK

⚠️ 1 exercise showing fatigue signal. Consider a deload week.
```

### 6. `hevy split [--weeks N]`
**What**: Actual training split analysis — what you're really doing vs what you think.
**Logic**:
1. Fetch last N weeks of workouts (default 4)
2. Categorize each workout by its dominant muscle groups (using exercise catalog)
3. Count frequency of each split type (Push/Pull/Legs/Upper/Lower/Full)
4. Show distribution and flag imbalances

**Workout categorization**: If >50% of exercises target chest/shoulders/triceps → Push. If >50% target back/biceps → Pull. If >50% target quads/hams/glutes/calves → Legs. Mixed → Upper/Lower/Full based on muscle mix.

**Output**:
```
Training Split (last 4 weeks)
TYPE     COUNT  PERCENTAGE  DAYS
Push         4       31%    Mon, Mon, Mon, Tue
Pull         4       31%    Wed, Wed, Thu, Wed  
Legs         3       23%    Fri, Fri, Mon
Upper        2       15%    Tue, Tue

⚠️ Legs slightly under-represented (23% vs 31% push). Consider adding a leg session.
```

**Flags**: `--weeks N` (default 4)

### 7. `hevy records`
**What**: All-time records vs current bests — shows regression or progression.
**Logic**:
1. Fetch all workouts
2. For each exercise, find: all-time best weight × reps, AND most recent best
3. Compare to show if you've regressed, maintained, or improved
4. Calculate estimated 1RM using Epley formula: `weight × (1 + reps/30)`

**Output**:
```
EXERCISE                   ALL-TIME BEST    CURRENT BEST    Δ        EST 1RM
Squat (Barbell)            80.00×8 (Feb)    70.00×12 (Mar)  -10.0kg  93.3kg
Bench Press (Barbell)      55.00×8 (Mar)    50.00×12 (Mar)  -5.0kg   70.0kg
Lat Pulldown (Cable)       45.00×12 (Mar)   45.00×12 (Mar)  =        63.0kg
```

**Flags**: `--top N` (show top N exercises by volume, default all)

### 8. `hevy rest`
**What**: Estimate rest periods between sets using workout timestamps.
**Logic**:
1. This is TRICKY — Hevy API sets don't have individual timestamps. We only have workout start_time and end_time.
2. Alternative approach: calculate average time per set = (end_time - start_time) / total_sets. Compare across workouts.
3. Show which workouts are most/least time-efficient (volume per minute)

**Output**:
```
Time Efficiency (last 10 sessions)
WORKOUT                    DATE       DURATION  SETS  VOL/MIN  EFFICIENCY
Upper                      Mar 31     75min     21    261kg    ████████
Legs                       Mar 30     68min     15    503kg    ██████████████
Day 1 - Upper A (Push)     Mar 25     82min     24    220kg    ██████
```

**Flags**: `--limit N` (default 10)

### 9. `hevy readiness`
**What**: WHOOP recovery + training recommendation. THE killer feature.
**Logic**:
1. Call the WHOOP Python script to get today's recovery: `python3 <whoop-skill-dir>/scripts/get_recovery.py --today --json`
   - WHOOP skill dir: stored in config as `whoop_script` or default to `~/.openclaw/workspace/skills/whoop-tracker`
   - If WHOOP not configured or script fails, print: "WHOOP not configured. Run: hevy config set whoop_path <path-to-whoop-skill>"
2. Parse recovery_score, hrv_rmssd, resting_hr
3. Cross-reference with planned workout:
   - Green (recovery ≥ 67%): "Full send. Heavy compounds OK."
   - Yellow (34-66%): "Moderate session. Reduce volume 20%, skip heavy singles."
   - Red (<34%): "Active recovery only. Light cardio or mobility."
4. Look at last workout's muscle groups to suggest what NOT to hit today (too soon)
5. If `hevy plan` logic suggests a routine, include it

**Output**:
```
🟢 WHOOP Recovery: 89%  |  HRV: 78ms  |  RHR: 56bpm

Status: GREEN — full send today
Last trained: Upper (yesterday)
Suggested: Legs or Push

Recovery history (last 7 days):
Mon  89%  ████████▉
Sun  72%  ███████▏
Sat  65%  ██████▌
Fri  81%  ████████
```

**Implementation notes for WHOOP integration**:
- Use `exec.Command("python3", scriptPath, "--today", "--json")` to call the WHOOP script
- The `get_recovery.py --today --json` outputs JSON like:
  ```json
  {
    "records": [{
      "score_state": "SCORED",
      "score": {
        "recovery_score": 89.0,
        "resting_heart_rate": 56.0,
        "hrv_rmssd_milli": 78.10088,
        "spo2_percentage": 94.125,
        "skin_temp_celsius": 33.352665
      }
    }],
    "next_token": null
  }
  ```
  Parse `records[0].score.recovery_score`, `.hrv_rmssd_milli`, `.resting_heart_rate`
- For the 7-day history, call `get_recovery.py --days 7 --json` — same structure, multiple records
- Store the WHOOP script path in the hevy config file (same config system as `hevy config set unit kg`)
- Config key: `whoop_path` (default: `~/.openclaw/workspace/skills/whoop-tracker`)

**Flags**: `--no-whoop` (skip WHOOP, just show training readiness based on workout history), `--days N` (recovery history days, default 7)

## Testing Requirements
1. `go build` must succeed with zero errors
2. Each command must work with `--json` flag
3. Each command must handle edge cases:
   - No workouts found
   - Single workout only
   - No RPE data (for fatigue)
   - No supersets (for supersets)
   - WHOOP not configured (for readiness)
4. Run `go vet ./...` — zero issues

## Build & Commit
```bash
go build -o ~/.local/bin/hevy .
go vet ./...
git add -A && git commit -m "feat(analytics): add phase 3 — plan, consistency, plateau, supersets, fatigue, split, records, rest, readiness

9 new commands:
- hevy plan: suggest next workout based on muscle recovery
- hevy consistency: monthly training consistency report  
- hevy plateau: detect stalled exercises
- hevy supersets: show superset pairings from history
- hevy fatigue: RPE trend analysis for accumulated fatigue
- hevy split: actual training split distribution
- hevy records: all-time vs current bests with est. 1RM
- hevy rest: time efficiency analysis per session
- hevy readiness: WHOOP recovery integration + training advice"
```

## File Organization
- Add all new commands to `cmd/insights.go` (or create `cmd/phase3.go` if insights.go gets too large — if you do, make sure init() registers all commands on rootCmd)
- Add any new analytics helper functions to `cmd/analytics.go`
- WHOOP integration helper can go in a new `cmd/whoop.go` file
- Keep the WHOOP dependency completely optional — readiness command must work gracefully without it
