package cmd

import (
	"testing"
	"time"

	"github.com/dhruvkelawala/hevy-cli/internal/api"
)

func TestBuildPlateauReportCountsCurrentPlateauWindow(t *testing.T) {
	w := func(v float64) *float64 { return &v }
	r := func(v int) *int { return &v }
	workouts := []api.Workout{
		{ID: "1", StartTime: "2026-03-01T10:00:00Z", Exercises: []api.Exercise{{Title: "Bench Press", Sets: []api.Set{{Type: "normal", WeightKG: w(45), Reps: r(5)}}}}},
		{ID: "2", StartTime: "2026-03-08T10:00:00Z", Exercises: []api.Exercise{{Title: "Bench Press", Sets: []api.Set{{Type: "normal", WeightKG: w(50), Reps: r(5)}}}}},
		{ID: "3", StartTime: "2026-03-15T10:00:00Z", Exercises: []api.Exercise{{Title: "Bench Press", Sets: []api.Set{{Type: "normal", WeightKG: w(50), Reps: r(5)}}}}},
		{ID: "4", StartTime: "2026-03-22T10:00:00Z", Exercises: []api.Exercise{{Title: "Bench Press", Sets: []api.Set{{Type: "normal", WeightKG: w(50), Reps: r(5)}}}}},
	}

	report := buildPlateauReport(workouts, 3, false)
	if len(report.Exercises) != 1 {
		t.Fatalf("expected 1 plateau entry, got %d", len(report.Exercises))
	}
	entry := report.Exercises[0]
	if entry.SessionsWithoutPR != 3 {
		t.Fatalf("expected plateau session count 3, got %d", entry.SessionsWithoutPR)
	}
	if !entry.Plateaued {
		t.Fatal("expected entry to be plateaued")
	}
	if got := entry.StalledSince.Format("2006-01-02"); got != "2026-03-08" {
		t.Fatalf("expected stalled since 2026-03-08, got %s", got)
	}
}

func TestBuildSupersetPairs(t *testing.T) {
	id := 1
	workouts := []api.Workout{{
		StartTime: "2026-03-31T10:00:00Z",
		Exercises: []api.Exercise{{Title: "Chest Fly", SupersetID: &id}, {Title: "Triceps Pushdown", SupersetID: &id}},
	}, {
		StartTime: "2026-03-29T10:00:00Z",
		Exercises: []api.Exercise{{Title: "Chest Fly", SupersetID: &id}, {Title: "Triceps Pushdown", SupersetID: &id}},
	}}
	pairs := buildSupersetPairs(workouts)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].Pair != "Chest Fly + Triceps Pushdown" || pairs[0].Count != 2 {
		t.Fatalf("unexpected pair: %#v", pairs[0])
	}
}

func TestReadinessWithoutWhoopRecentWorkout(t *testing.T) {
	now := time.Date(2026, time.March, 31, 12, 0, 0, 0, time.UTC)
	last := now.Add(-24 * time.Hour)
	status, advice := readinessWithoutWhoop([]string{"chest", "triceps"}, &last, now)
	if status != "CAUTION" {
		t.Fatalf("expected CAUTION, got %s", status)
	}
	if advice == "" {
		t.Fatal("expected advice")
	}
}
