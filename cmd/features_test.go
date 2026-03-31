package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dhruvkelawala/hevy-cli/internal/api"
	"github.com/spf13/cobra"
)

func TestDetailCommandsAdvertiseRequiredIDs(t *testing.T) {
	if routineCmd.Use != "routine <routine-id>" {
		t.Fatalf("unexpected routine usage: %q", routineCmd.Use)
	}
	if exerciseCmd.Use != "exercise <exercise-id>" {
		t.Fatalf("unexpected exercise usage: %q", exerciseCmd.Use)
	}
	if historyCmd.Use != "history <exercise-template-id>" {
		t.Fatalf("unexpected history usage: %q", historyCmd.Use)
	}
	if !strings.Contains(historyCmd.Long, "hevy exercises") {
		t.Fatalf("expected history help to explain ID source, got %q", historyCmd.Long)
	}
	if !strings.Contains(historyCmd.Example, "hevy history 7EB3F7C3") {
		t.Fatalf("expected history example to stay accurate, got %q", historyCmd.Example)
	}
}

func TestRequireSingleIdentifierArg(t *testing.T) {
	root := &cobra.Command{Use: "hevy"}
	cmd := &cobra.Command{Use: "routine <routine-id>"}
	root.AddCommand(cmd)

	if err := requireSingleIdentifierArg("routine-id")(cmd, []string{"abc123"}); err != nil {
		t.Fatalf("expected single arg to pass, got %v", err)
	}

	err := requireSingleIdentifierArg("routine-id")(cmd, nil)
	if err == nil {
		t.Fatal("expected missing arg error")
	}
	if !strings.Contains(err.Error(), "requires exactly 1 argument: <routine-id>") {
		t.Fatalf("unexpected arg error: %v", err)
	}
	if !strings.Contains(err.Error(), "Usage: hevy routine <routine-id>") {
		t.Fatalf("expected usage line in error, got %v", err)
	}
}

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

func TestNeedsExerciseCatalog(t *testing.T) {
	if needsExerciseCatalog("", "", false) {
		t.Fatal("expected paginated listing without filters")
	}
	if !needsExerciseCatalog("bench", "", false) {
		t.Fatal("expected search to require full exercise catalog")
	}
	if !needsExerciseCatalog("", "chest", false) {
		t.Fatal("expected muscle filter to require full exercise catalog")
	}
	if !needsExerciseCatalog("", "", true) {
		t.Fatal("expected custom filter to require full exercise catalog")
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

func TestPickExerciseByNamePrefersPrefixAndBetterMatches(t *testing.T) {
	exercises := []api.ExerciseTemplate{
		{ID: "1", Title: "Assisted Pistol Squats"},
		{ID: "2", Title: "Squat (Band)"},
		{ID: "3", Title: "Squat (Barbell)"},
		{ID: "4", Title: "Hack Squat"},
	}

	match := pickExerciseByName(exercises, "Squat")
	if match == nil {
		t.Fatal("expected a match for Squat")
	}
	if match.ID != "3" {
		t.Fatalf("expected Squat (Barbell), got %#v", match)
	}
}

func TestCalculateWorkoutStreak(t *testing.T) {
	workouts := []api.Workout{
		{StartTime: "2026-03-30T10:00:00Z"},
		{StartTime: "2026-03-24T10:00:00Z"},
		{StartTime: "2026-03-17T10:00:00Z"},
		{StartTime: "2026-03-10T10:00:00Z"},
		{StartTime: "2026-02-20T10:00:00Z"},
	}

	streak, since := calculateWorkoutStreak(workouts)
	if streak != 4 {
		t.Fatalf("expected 4-week streak, got %d", streak)
	}
	if since.Format("2006-01-02") != "2026-03-09" {
		t.Fatalf("expected streak start 2026-03-09, got %s", since.Format("2006-01-02"))
	}
}

func TestFindPersonalRecord(t *testing.T) {
	w100 := 100.0
	w105 := 105.0
	r5 := 5
	r3 := 3
	record, ok := findPersonalRecord("Bench Press", []api.ExerciseHistoryEntry{
		{WorkoutStartTime: "2026-03-10T10:00:00Z", WeightKG: &w100, Reps: &r5},
		{WorkoutStartTime: "2026-03-17T10:00:00Z", WeightKG: &w105, Reps: &r3},
	})
	if !ok {
		t.Fatal("expected PR to be found")
	}
	if record.WeightKG != 105 || record.Reps != 3 || record.Exercise != "Bench Press" {
		t.Fatalf("unexpected record: %#v", record)
	}
}

func TestBuildVolumePointsAndChart(t *testing.T) {
	app.weightUnit = "kg"
	w1 := 100.0
	w2 := 110.0
	r5 := 5
	r3 := 3
	entries := []api.ExerciseHistoryEntry{
		{WorkoutID: "w1", WorkoutStartTime: "2026-03-10T10:00:00Z", WeightKG: &w1, Reps: &r5},
		{WorkoutID: "w1", WorkoutStartTime: "2026-03-10T10:00:00Z", WeightKG: &w1, Reps: &r3},
		{WorkoutID: "w2", WorkoutStartTime: "2026-03-17T10:00:00Z", WeightKG: &w2, Reps: &r5},
	}
	points := buildVolumePoints(entries)
	if len(points) != 2 {
		t.Fatalf("expected 2 volume points, got %d", len(points))
	}
	if points[0].VolumeKG != 800 || points[1].VolumeKG != 550 {
		t.Fatalf("unexpected point volumes: %#v", points)
	}
	chart := renderVolumeChart("Bench Press", points, 8)
	if !strings.Contains(chart[0], "Bench Press volume") {
		t.Fatalf("unexpected chart title: %#v", chart)
	}
	if !strings.Contains(chart[1], "800.00") {
		t.Fatalf("expected first chart row to include volume, got %#v", chart)
	}
	if !strings.Contains(chart[2], "550.00") {
		t.Fatalf("expected second chart row to include volume, got %#v", chart)
	}
	if !strings.Contains(chart[1], "█") || !strings.Contains(chart[2], "█") {
		t.Fatalf("expected chart bars, got %#v", chart)
	}
}

func TestHistoryEmptyStateLines(t *testing.T) {
	lines := historyEmptyStateLines("3BC06AD3", true)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "No logged history found for exercise template ID 3BC06AD3.") {
		t.Fatalf("unexpected empty state: %q", joined)
	}
	if !strings.Contains(joined, "hevy exercises") {
		t.Fatalf("expected ID source hint, got %q", joined)
	}
	if !strings.Contains(joined, "--start-date/--end-date") {
		t.Fatalf("expected filtered history tip, got %q", joined)
	}
}

func TestRenderExerciseHistoryEmptyState(t *testing.T) {
	originalMode := app.outputMode
	defer func() { app.outputMode = originalMode }()
	app.outputMode = outputTable

	buf := &bytes.Buffer{}
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = originalStdout }()

	done := make(chan error, 1)
	go func() {
		_, copyErr := buf.ReadFrom(r)
		done <- copyErr
	}()

	if err := renderExerciseHistory("3BC06AD3", &api.ExerciseHistoryResponse{}, false); err != nil {
		t.Fatalf("renderExerciseHistory returned error: %v", err)
	}
	_ = w.Close()
	if err := <-done; err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	_ = r.Close()

	output := buf.String()
	if !strings.Contains(output, "No logged history found for exercise template ID 3BC06AD3.") {
		t.Fatalf("unexpected render output: %q", output)
	}
	if !strings.Contains(output, "exercise_template_id") {
		t.Fatalf("expected ID source guidance, got %q", output)
	}
}

func TestRenderCalendar(t *testing.T) {
	lines := renderCalendar(2026, time.March, map[int]bool{9: true, 12: true, 16: true, 18: true, 19: true, 25: true, 27: true, 30: true})
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "March 2026") {
		t.Fatalf("expected month heading, got %q", joined)
	}
	if !strings.Contains(joined, "[9]") || !strings.Contains(joined, "[12]") || !strings.Contains(joined, "[30]") {
		t.Fatalf("expected workout days to be highlighted, got %q", joined)
	}
	if !strings.Contains(joined, "Mo Tu We Th Fr Sa Su") {
		t.Fatalf("expected weekday heading, got %q", joined)
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
