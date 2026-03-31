package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/dhruvkelawala/go-hevy/internal/api"
)

func TestFilterExercises(t *testing.T) {
	exercises := []api.ExerciseTemplate{
		{ID: "1", Title: "Bench Press", PrimaryMuscleGroup: "chest", SecondaryMuscleGroups: []string{"triceps"}},
		{ID: "2", Title: "Incline Dumbbell Press", PrimaryMuscleGroup: "chest", SecondaryMuscleGroups: []string{"shoulders"}, IsCustom: true},
		{ID: "3", Title: "Barbell Row", PrimaryMuscleGroup: "back", SecondaryMuscleGroups: []string{"biceps"}},
	}

	filtered := filterExercises(exercises, "press", "chest", false)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 chest press exercises, got %d", len(filtered))
	}

	filtered = filterExercises(exercises, "", "shoulders", false)
	if len(filtered) != 1 || filtered[0].ID != "2" {
		t.Fatalf("expected shoulder secondary match, got %#v", filtered)
	}

	filtered = filterExercises(exercises, "", "", true)
	if len(filtered) != 1 || filtered[0].ID != "2" {
		t.Fatalf("expected only custom exercise, got %#v", filtered)
	}
}

func TestWorkoutCSVRows(t *testing.T) {
	weight := 100.0
	reps := 5
	rpe := 8.5
	workouts := []api.Workout{{
		StartTime: "2026-03-31T10:00:00Z",
		Title:     "Push Day",
		Exercises: []api.Exercise{{
			Title: "Bench Press",
			Sets:  []api.Set{{Index: 1, Type: "normal", WeightKG: &weight, Reps: &reps, RPE: &rpe}},
		}},
	}}

	rows := workoutCSVRows(workouts)
	if len(rows) != 2 {
		t.Fatalf("expected header + 1 row, got %d rows", len(rows))
	}
	if rows[1][0] != "2026-03-31T10:00:00Z" || rows[1][1] != "Push Day" || rows[1][2] != "Bench Press" {
		t.Fatalf("unexpected csv row: %#v", rows[1])
	}

	encoded, err := encodeCSV(rows)
	if err != nil {
		t.Fatalf("encodeCSV returned error: %v", err)
	}
	if !strings.Contains(encoded, "date,workout_title,exercise,set_number,type,weight_kg,reps,rpe") {
		t.Fatalf("expected CSV header, got %q", encoded)
	}
	if !strings.Contains(encoded, "Push Day,Bench Press,1,normal,100,5,8.5") {
		t.Fatalf("expected CSV row contents, got %q", encoded)
	}
}

func TestBuildProgressPointsAndChart(t *testing.T) {
	app.weightUnit = "kg"
	w1a := 60.0
	w1b := 62.5
	w2 := 65.0
	entries := []api.ExerciseHistoryEntry{
		{WorkoutID: "w1", WorkoutStartTime: mustTime(t, "2026-03-12T10:00:00Z").Format(time.RFC3339), WeightKG: &w1a},
		{WorkoutID: "w1", WorkoutStartTime: mustTime(t, "2026-03-12T10:00:00Z").Format(time.RFC3339), WeightKG: &w1b},
		{WorkoutID: "w2", WorkoutStartTime: mustTime(t, "2026-03-15T10:00:00Z").Format(time.RFC3339), WeightKG: &w2},
	}

	points := buildProgressPoints(entries)
	if len(points) != 2 {
		t.Fatalf("expected 2 progress points, got %d", len(points))
	}
	if points[0].Weight != 62.5 || points[1].Weight != 65.0 {
		t.Fatalf("expected max workout weights, got %#v", points)
	}

	chart := renderProgressChart("Bench Press", points, 8)
	if len(chart) != 3 {
		t.Fatalf("expected title + 2 chart rows, got %d", len(chart))
	}
	if !strings.Contains(chart[0], "Bench Press — last 2 sessions") {
		t.Fatalf("unexpected chart title: %q", chart[0])
	}
	if !strings.Contains(chart[1], "62.50kg") || !strings.Contains(chart[2], "65.00kg") {
		t.Fatalf("unexpected chart body: %#v", chart)
	}
	if !strings.Contains(chart[1], "█") || !strings.Contains(chart[2], "█") {
		t.Fatalf("expected bars in chart, got %#v", chart)
	}
}

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return parsed
}
