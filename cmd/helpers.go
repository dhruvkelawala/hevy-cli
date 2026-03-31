package cmd

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dhruvkelawala/hevy-cli/internal/api"
	appconfig "github.com/dhruvkelawala/hevy-cli/internal/config"
	"github.com/dhruvkelawala/hevy-cli/internal/output"
	"github.com/fatih/color"
)

const poundsPerKilogram = 2.2046226218

func requirePositivePagination(page, pageSize, maxPageSize int) error {
	if page < 1 {
		return fmt.Errorf("page must be 1 or greater")
	}
	if pageSize < 1 || pageSize > maxPageSize {
		return fmt.Errorf("page-size must be between 1 and %d", maxPageSize)
	}
	return nil
}

func readWorkoutRequestFile(path string) (api.CreateWorkoutRequest, error) {
	var payload api.CreateWorkoutRequest
	data, err := os.ReadFile(path)
	if err != nil {
		return payload, fmt.Errorf("read file: %w", err)
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return payload, fmt.Errorf("parse JSON file: %w", err)
	}
	return payload, nil
}

func formatTimestamp(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return t.Local().Format("2006-01-02 15:04")
}

func formatFloatPtr(v *float64) string {
	if v == nil {
		return "-"
	}
	return formatWeight(*v)
}

func formatIntPtr(v *int) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *v)
}

func printObject(v any, tableRows [][2]string, compactLines []string) error {
	switch app.outputMode {
	case outputJSON:
		return output.PrintJSON(os.Stdout, v)
	case outputCompact:
		return output.PrintCompact(os.Stdout, compactLines)
	default:
		output.PrintKeyValueTable(os.Stdout, tableRows)
		return nil
	}
}

func configRows() [][2]string {
	path, _ := appconfig.ConfigPath()
	defaultLimit := ""
	if app.config.DefaultLimit > 0 {
		defaultLimit = strconv.Itoa(app.config.DefaultLimit)
	}
	return [][2]string{
		{"config_path", output.ValueOrDash(path)},
		{"api_key", output.ValueOrDash(appconfig.Redact(app.config.EffectiveAPIKey()))},
		{"unit", output.ValueOrDash(app.config.Unit)},
		{"default_limit", output.ValueOrDash(defaultLimit)},
		{"api_key_source", configSource()},
	}
}

func configSource() string {
	if strings.TrimSpace(os.Getenv(appconfig.EnvAPIKey)) != "" {
		return "environment"
	}
	if app.config != nil && app.config.HasAPIKey() {
		return "config_file"
	}
	return "unset"
}

func formatWeight(kg float64) string {
	if app.weightUnit == "lbs" {
		return fmt.Sprintf("%.1f", kg*poundsPerKilogram)
	}
	return fmt.Sprintf("%.2f", kg)
}

func weightHeader() string {
	if app.weightUnit == "lbs" {
		return "Weight LBS"
	}
	return "Weight KG"
}

func colorWorkoutTitle(title string) string {
	return color.New(color.Bold).Sprint(title)
}

func colorSetType(set api.Set, isPR bool) string {
	if isPR {
		return color.New(color.FgGreen).Sprint(set.Type)
	}
	switch strings.ToLower(set.Type) {
	case "warmup":
		return color.New(color.FgYellow).Sprint(set.Type)
	default:
		return color.New(color.FgWhite).Sprint(set.Type)
	}
}

func isPersonalRecord(weight *float64, bestSoFar float64) bool {
	return weight != nil && *weight > bestSoFar
}

func matchExercise(ex api.ExerciseTemplate, query, muscle string, customOnly bool) bool {
	if query != "" && !strings.Contains(strings.ToLower(ex.Title), strings.ToLower(query)) {
		return false
	}
	if muscle != "" {
		needle := strings.ToLower(strings.TrimSpace(muscle))
		if strings.ToLower(ex.PrimaryMuscleGroup) != needle {
			matched := false
			for _, group := range ex.SecondaryMuscleGroups {
				if strings.ToLower(group) == needle {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		}
	}
	if customOnly && !ex.IsCustom {
		return false
	}
	return true
}

func filterExercises(exercises []api.ExerciseTemplate, query, muscle string, customOnly bool) []api.ExerciseTemplate {
	filtered := make([]api.ExerciseTemplate, 0, len(exercises))
	for _, exercise := range exercises {
		if matchExercise(exercise, query, muscle, customOnly) {
			filtered = append(filtered, exercise)
		}
	}
	return filtered
}

func needsExerciseCatalog(query, muscle string, customOnly bool) bool {
	return strings.TrimSpace(query) != "" || strings.TrimSpace(muscle) != "" || customOnly
}

func fetchAllExercises(client *api.Client) ([]api.ExerciseTemplate, error) {
	page := 1
	all := []api.ExerciseTemplate{}
	for {
		resp, err := client.ListExercises(context.Background(), page, 100)
		if err != nil {
			return nil, err
		}
		all = append(all, resp.ExerciseTemplates...)
		if page >= resp.PageCount || len(resp.ExerciseTemplates) == 0 {
			break
		}
		page++
	}
	return all, nil
}

func fetchWorkouts(client *api.Client, startPage, limit int, fetchAll bool) ([]api.Workout, error) {
	if fetchAll {
		limit = 0
	}
	page := startPage
	workouts := []api.Workout{}
	for {
		pageSize := 10
		if !fetchAll && limit > 0 && limit-len(workouts) < pageSize {
			pageSize = limit - len(workouts)
		}
		if pageSize <= 0 {
			break
		}
		resp, err := client.ListWorkouts(context.Background(), page, pageSize)
		if err != nil {
			return nil, err
		}
		workouts = append(workouts, resp.Workouts...)
		if len(resp.Workouts) == 0 || page >= resp.PageCount || (!fetchAll && limit > 0 && len(workouts) >= limit) {
			break
		}
		page++
	}
	if !fetchAll && limit > 0 && len(workouts) > limit {
		workouts = workouts[:limit]
	}
	return workouts, nil
}

func fetchWorkoutDetails(client *api.Client, workouts []api.Workout) ([]api.Workout, error) {
	detailed := make([]api.Workout, 0, len(workouts))
	for _, workout := range workouts {
		fullWorkout, err := client.GetWorkout(context.Background(), workout.ID)
		if err != nil {
			return nil, err
		}
		detailed = append(detailed, *fullWorkout)
	}
	return detailed, nil
}

func exerciseCatalogMap(exercises []api.ExerciseTemplate) map[string]api.ExerciseTemplate {
	catalog := make(map[string]api.ExerciseTemplate, len(exercises))
	for _, exercise := range exercises {
		catalog[exercise.ID] = exercise
	}
	return catalog
}

func pickExerciseByName(exercises []api.ExerciseTemplate, name string) *api.ExerciseTemplate {
	needle := strings.ToLower(strings.TrimSpace(name))
	if needle == "" {
		return nil
	}
	bestScore := 0
	var best *api.ExerciseTemplate
	for i := range exercises {
		score := exerciseMatchScore(exercises[i].Title, needle)
		if score > bestScore {
			bestScore = score
			copy := exercises[i]
			best = &copy
		}
	}
	return best
}

func exerciseMatchScore(title, needle string) int {
	title = strings.ToLower(strings.TrimSpace(title))
	if title == needle {
		return 1000
	}

	score := 0
	if strings.HasPrefix(title, needle) {
		score = 800
	}
	if strings.Contains(title, needle) && score < 600 {
		score = 600
	}
	for _, word := range strings.Fields(title) {
		if strings.HasPrefix(word, needle) && score < 400 {
			score = 400
		}
	}
	if score == 0 {
		return 0
	}

	if strings.Contains(title, "(barbell)") {
		score += 50
	}
	if strings.Contains(title, "(dumbbell)") {
		score += 25
	}
	if strings.Contains(title, "assisted") || strings.Contains(title, "band") {
		score -= 25
	}

	return score
}

type progressPoint struct {
	Date   time.Time
	Label  string
	Weight float64
}

func buildProgressPoints(entries []api.ExerciseHistoryEntry) []progressPoint {
	bestByWorkout := map[string]progressPoint{}
	for _, entry := range entries {
		if entry.WeightKG == nil || strings.TrimSpace(entry.WorkoutStartTime) == "" {
			continue
		}
		date, err := time.Parse(time.RFC3339, entry.WorkoutStartTime)
		if err != nil {
			continue
		}
		current, ok := bestByWorkout[entry.WorkoutID]
		if !ok || *entry.WeightKG > current.Weight {
			bestByWorkout[entry.WorkoutID] = progressPoint{Date: date, Label: date.Format("Jan 02"), Weight: *entry.WeightKG}
		}
	}
	points := make([]progressPoint, 0, len(bestByWorkout))
	for _, point := range bestByWorkout {
		points = append(points, point)
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Date.Before(points[j].Date) })
	return points
}

func renderProgressChart(title string, points []progressPoint, maxPoints int) []string {
	if maxPoints > 0 && len(points) > maxPoints {
		points = points[len(points)-maxPoints:]
	}
	lines := []string{fmt.Sprintf("%s — last %d sessions", title, len(points))}
	if len(points) == 0 {
		return append(lines, "No weighted history found.")
	}
	maxWeight := 0.0
	for _, point := range points {
		if point.Weight > maxWeight {
			maxWeight = point.Weight
		}
	}
	for _, point := range points {
		width := 1
		if maxWeight > 0 {
			width = int((point.Weight / maxWeight) * 14)
			if width < 1 {
				width = 1
			}
		}
		lines = append(lines, fmt.Sprintf("%s  %s%s  %s", point.Label, formatWeight(point.Weight), app.weightUnit, strings.Repeat("█", width)))
	}
	return lines
}

func workoutCSVRows(workouts []api.Workout) [][]string {
	rows := [][]string{{"date", "workout_title", "exercise", "set_number", "type", "weight_kg", "reps", "rpe"}}
	for _, workout := range workouts {
		for _, exercise := range workout.Exercises {
			for _, set := range exercise.Sets {
				rows = append(rows, []string{
					workout.StartTime,
					workout.Title,
					exercise.Title,
					strconv.Itoa(set.Index),
					set.Type,
					csvFloat(set.WeightKG),
					csvInt(set.Reps),
					csvFloat(set.RPE),
				})
			}
		}
	}
	return rows
}

func encodeCSV(rows [][]string) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.WriteAll(rows); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func csvFloat(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}

func csvInt(v *int) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(*v)
}
