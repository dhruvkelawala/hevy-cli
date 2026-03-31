package cmd

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dhruvkelawala/hevy-cli/internal/api"
)

type planRecommendation struct {
	LookbackDays       int                 `json:"lookback_days"`
	LastTrained        []muscleLastTrained `json:"last_trained"`
	SuggestedRoutine   string              `json:"suggested_routine,omitempty"`
	SuggestedRoutineID string              `json:"suggested_routine_id,omitempty"`
	RoutineScore       float64             `json:"routine_score"`
	Reason             string              `json:"reason,omitempty"`
	RoutineMuscles     []string            `json:"routine_muscles,omitempty"`
}

type muscleLastTrained struct {
	Muscle      string    `json:"muscle"`
	LastTrained time.Time `json:"last_trained"`
	DaysAgo     int       `json:"days_ago"`
	Weekday     string    `json:"weekday"`
}

type monthlyConsistency struct {
	Month          string  `json:"month"`
	Sessions       int     `json:"sessions"`
	Target         int     `json:"target"`
	HitRate        float64 `json:"hit_rate"`
	TotalVolumeKG  float64 `json:"total_volume_kg"`
	LongestGapDays int     `json:"longest_gap_days"`
}

type consistencyReport struct {
	Months              int                  `json:"months"`
	TargetPerWeek       int                  `json:"target_per_week"`
	Monthly             []monthlyConsistency `json:"monthly"`
	OverallSessions     int                  `json:"overall_sessions"`
	OverallTarget       int                  `json:"overall_target"`
	OverallHitRate      float64              `json:"overall_hit_rate"`
	CurrentTargetStreak int                  `json:"current_target_streak_weeks"`
}

type plateauEntry struct {
	Exercise          string    `json:"exercise"`
	CurrentWeightKG   float64   `json:"current_weight_kg"`
	StalledSince      time.Time `json:"stalled_since"`
	SessionsWithoutPR int       `json:"sessions_without_pr"`
	Plateaued         bool      `json:"plateaued"`
}

type plateauReport struct {
	Threshold int            `json:"threshold"`
	Exercises []plateauEntry `json:"exercises"`
}

type supersetPair struct {
	Pair     string    `json:"pair"`
	Count    int       `json:"count"`
	LastUsed time.Time `json:"last_used"`
}

type fatigueEntry struct {
	Exercise      string    `json:"exercise"`
	RPETrend      []float64 `json:"rpe_trend"`
	WeightTrendKG []float64 `json:"weight_trend_kg"`
	Signal        string    `json:"signal"`
}

type fatigueReport struct {
	Exercises []fatigueEntry `json:"exercises"`
	Signals   int            `json:"signals"`
	Message   string         `json:"message,omitempty"`
}

type splitEntry struct {
	Type       string   `json:"type"`
	Count      int      `json:"count"`
	Percentage float64  `json:"percentage"`
	Days       []string `json:"days"`
}

type splitReport struct {
	Weeks   int          `json:"weeks"`
	Total   int          `json:"total_workouts"`
	Entries []splitEntry `json:"entries"`
	Warning string       `json:"warning,omitempty"`
}

type recordEntry struct {
	Exercise         string    `json:"exercise"`
	AllTimeWeightKG  float64   `json:"all_time_weight_kg"`
	AllTimeReps      int       `json:"all_time_reps"`
	AllTimeDate      time.Time `json:"all_time_date"`
	CurrentWeightKG  float64   `json:"current_weight_kg"`
	CurrentReps      int       `json:"current_reps"`
	CurrentDate      time.Time `json:"current_date"`
	DeltaWeightKG    float64   `json:"delta_weight_kg"`
	EstimatedOneRMKG float64   `json:"estimated_one_rm_kg"`
	LifetimeVolumeKG float64   `json:"lifetime_volume_kg"`
}

type restEntry struct {
	Workout           string    `json:"workout"`
	Date              time.Time `json:"date"`
	DurationMinutes   float64   `json:"duration_minutes"`
	SetCount          int       `json:"set_count"`
	VolumePerMinuteKG float64   `json:"volume_per_minute_kg"`
	AverageSetMinutes float64   `json:"average_set_minutes"`
}

type readinessReport struct {
	Whoop            *whoopRecoverySnapshot `json:"whoop,omitempty"`
	RecoveryHistory  []whoopHistoryPoint    `json:"recovery_history,omitempty"`
	Status           string                 `json:"status"`
	Advice           string                 `json:"advice"`
	AvoidMuscles     []string               `json:"avoid_muscles,omitempty"`
	LastWorkoutTitle string                 `json:"last_workout_title,omitempty"`
	LastWorkoutDate  *time.Time             `json:"last_workout_date,omitempty"`
	Plan             *planRecommendation    `json:"plan,omitempty"`
	WhoopMessage     string                 `json:"whoop_message,omitempty"`
}

type exerciseSessionPoint struct {
	Date       time.Time
	WeightKG   float64
	AverageRPE float64
	HasRPE     bool
	WorkoutID  string
	VolumeKG   float64
	Reps       int
}

type workoutSlice struct {
	Workout api.Workout
	Start   time.Time
}

func filterWorkoutsSince(workouts []api.Workout, since time.Time) []api.Workout {
	filtered := make([]api.Workout, 0, len(workouts))
	for _, workout := range workouts {
		startedAt, ok := parseAPITime(workout.StartTime)
		if !ok || startedAt.Before(since) {
			continue
		}
		filtered = append(filtered, workout)
	}
	return filtered
}

func sortWorkoutsChronologically(workouts []api.Workout) []workoutSlice {
	items := make([]workoutSlice, 0, len(workouts))
	for _, workout := range workouts {
		startedAt, ok := parseAPITime(workout.StartTime)
		if !ok {
			continue
		}
		items = append(items, workoutSlice{Workout: workout, Start: startedAt})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Start.Before(items[j].Start) })
	return items
}

func exerciseMuscles(exercise api.Exercise, catalog map[string]api.ExerciseTemplate) []string {
	if template, ok := catalog[exercise.ExerciseTemplateID]; ok {
		muscles := make([]string, 0, 1+len(template.SecondaryMuscleGroups))
		if primary := normalizeMuscle(template.PrimaryMuscleGroup); primary != "" {
			muscles = append(muscles, primary)
		}
		for _, secondary := range template.SecondaryMuscleGroups {
			if muscle := normalizeMuscle(secondary); muscle != "" && !containsString(muscles, muscle) {
				muscles = append(muscles, muscle)
			}
		}
		if len(muscles) > 0 {
			return muscles
		}
	}
	primary := normalizeMuscle(primaryMuscleForExercise(exercise, catalog))
	if primary == "" {
		return nil
	}
	return []string{primary}
}

func normalizeMuscle(value string) string {
	muscle := strings.ToLower(strings.TrimSpace(value))
	muscle = strings.ReplaceAll(muscle, " ", "_")
	switch muscle {
	case "lats", "upper_back", "lower_back", "middle_back", "traps":
		return "back"
	case "quadriceps", "quads":
		return "quads"
	case "hamstrings", "hams":
		return "hamstrings"
	case "gluteus", "glutes":
		return "glutes"
	case "calf", "calves":
		return "calves"
	case "tricep", "triceps":
		return "triceps"
	case "bicep", "biceps", "forearms":
		return "biceps"
	case "delts", "rear_delts", "front_delts", "side_delts":
		return "shoulders"
	case "legs", "adductors", "abductors":
		return "legs"
	default:
		return muscle
	}
}

func muscleBucket(muscle string) string {
	switch normalizeMuscle(muscle) {
	case "chest", "shoulders", "triceps":
		return "push"
	case "back", "biceps":
		return "pull"
	case "legs", "quads", "hamstrings", "glutes", "calves":
		return "legs"
	default:
		return "other"
	}
}

func buildLastTrained(workouts []api.Workout, catalog map[string]api.ExerciseTemplate, now time.Time) []muscleLastTrained {
	latest := map[string]time.Time{}
	for _, item := range sortWorkoutsChronologically(workouts) {
		seen := map[string]bool{}
		for _, exercise := range item.Workout.Exercises {
			for _, muscle := range exerciseMuscles(exercise, catalog) {
				if seen[muscle] {
					continue
				}
				seen[muscle] = true
				latest[muscle] = item.Start
			}
		}
	}
	entries := make([]muscleLastTrained, 0, len(latest))
	for muscle, date := range latest {
		daysAgo := int(now.Sub(date).Hours() / 24)
		entries = append(entries, muscleLastTrained{Muscle: muscle, LastTrained: date, DaysAgo: daysAgo, Weekday: date.Format("Mon")})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].DaysAgo == entries[j].DaysAgo {
			return entries[i].Muscle < entries[j].Muscle
		}
		return entries[i].DaysAgo > entries[j].DaysAgo
	})
	return entries
}

func recommendRoutine(workouts []api.Workout, routines []api.Routine, catalog map[string]api.ExerciseTemplate, lookbackDays int, now time.Time) planRecommendation {
	lastTrained := buildLastTrained(workouts, catalog, now)
	recommendation := planRecommendation{LookbackDays: lookbackDays, LastTrained: lastTrained}
	if len(routines) == 0 {
		recommendation.Reason = "No routines found to score."
		return recommendation
	}
	lastMap := map[string]muscleLastTrained{}
	for _, entry := range lastTrained {
		lastMap[entry.Muscle] = entry
	}
	bestScore := -1.0
	bestReason := ""
	for _, routine := range routines {
		muscles := musclesForRoutine(routine, catalog)
		if len(muscles) == 0 {
			continue
		}
		score := 0.0
		due := make([]muscleLastTrained, 0, len(muscles))
		for _, muscle := range muscles {
			if entry, ok := lastMap[muscle]; ok {
				score += float64(entry.DaysAgo)
				due = append(due, entry)
				continue
			}
			score += float64(lookbackDays + 1)
			due = append(due, muscleLastTrained{Muscle: muscle, DaysAgo: lookbackDays + 1})
		}
		score /= float64(len(muscles))
		sort.Slice(due, func(i, j int) bool { return due[i].DaysAgo > due[j].DaysAgo })
		reasonParts := make([]string, 0, minInt(2, len(due)))
		for _, item := range due[:minInt(2, len(due))] {
			if item.DaysAgo > lookbackDays {
				reasonParts = append(reasonParts, fmt.Sprintf("%s not hit in the last %d days", strings.ReplaceAll(item.Muscle, "_", " "), lookbackDays))
			} else {
				reasonParts = append(reasonParts, fmt.Sprintf("%s hasn't been hit in %d days", strings.ReplaceAll(item.Muscle, "_", " "), item.DaysAgo))
			}
		}
		if score > bestScore {
			bestScore = score
			recommendation.SuggestedRoutine = routine.Title
			recommendation.SuggestedRoutineID = routine.ID
			recommendation.RoutineScore = score
			recommendation.RoutineMuscles = muscles
			bestReason = strings.Join(reasonParts, ", ")
		}
	}
	if recommendation.SuggestedRoutine == "" {
		recommendation.Reason = "No routines with identifiable muscle groups found."
		return recommendation
	}
	if bestReason == "" {
		bestReason = "Best match based on recovery window and training frequency."
	}
	recommendation.Reason = bestReason
	return recommendation
}

func musclesForRoutine(routine api.Routine, catalog map[string]api.ExerciseTemplate) []string {
	muscles := []string{}
	for _, exercise := range routine.Exercises {
		for _, muscle := range exerciseMuscles(exercise.Exercise, catalog) {
			if !containsString(muscles, muscle) {
				muscles = append(muscles, muscle)
			}
		}
	}
	sort.Strings(muscles)
	return muscles
}

func buildConsistencyReport(workouts []api.Workout, months, target int, now time.Time) consistencyReport {
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -(months - 1), 0)
	filtered := filterWorkoutsSince(workouts, start)
	monthlyMap := map[string]*monthlyConsistency{}
	monthlyDays := map[string][]time.Time{}
	weeklySessions := map[time.Time]int{}
	for _, item := range sortWorkoutsChronologically(filtered) {
		monthKey := item.Start.Format("2006-01")
		entry := monthlyMap[monthKey]
		if entry == nil {
			entry = &monthlyConsistency{Month: item.Start.Format("Jan 2006")}
			monthlyMap[monthKey] = entry
		}
		entry.Sessions++
		entry.TotalVolumeKG += workoutVolume(item.Workout)
		monthlyDays[monthKey] = append(monthlyDays[monthKey], item.Start)
		weeklySessions[weekStart(item.Start)]++
	}
	report := consistencyReport{Months: months, TargetPerWeek: target}
	for i := 0; i < months; i++ {
		month := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location()).AddDate(0, i, 0)
		key := month.Format("2006-01")
		entry := monthlyMap[key]
		if entry == nil {
			entry = &monthlyConsistency{Month: month.Format("Jan 2006")}
		}
		entry.Target = countTargetWeeksInMonth(month) * target
		entry.LongestGapDays = longestGapDays(monthlyDays[key])
		if entry.Target > 0 {
			entry.HitRate = float64(entry.Sessions) / float64(entry.Target)
		}
		report.Monthly = append(report.Monthly, *entry)
		report.OverallSessions += entry.Sessions
		report.OverallTarget += entry.Target
	}
	if report.OverallTarget > 0 {
		report.OverallHitRate = float64(report.OverallSessions) / float64(report.OverallTarget)
	}
	weeks := make([]time.Time, 0, len(weeklySessions))
	for week := range weeklySessions {
		weeks = append(weeks, week)
	}
	sort.Slice(weeks, func(i, j int) bool { return weeks[i].After(weeks[j]) })
	for _, week := range weeks {
		if weeklySessions[week] >= target {
			report.CurrentTargetStreak++
			continue
		}
		break
	}
	return report
}

func countTargetWeeksInMonth(month time.Time) int {
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	last := first.AddDate(0, 1, -1)
	count := 0
	for current := weekStart(first); !current.After(last); current = current.AddDate(0, 0, 7) {
		days := 0
		for d := 0; d < 7; d++ {
			candidate := current.AddDate(0, 0, d)
			if candidate.Month() == month.Month() {
				days++
			}
		}
		if days >= 4 {
			count++
		}
	}
	if count == 0 {
		return 1
	}
	return count
}

func longestGapDays(days []time.Time) int {
	if len(days) < 2 {
		return 0
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })
	longest := 0
	for i := 1; i < len(days); i++ {
		gap := int(days[i].Sub(days[i-1]).Hours() / 24)
		if gap > longest {
			longest = gap
		}
	}
	return longest
}

func buildExerciseSessions(workouts []api.Workout) map[string][]exerciseSessionPoint {
	sessions := map[string][]exerciseSessionPoint{}
	for _, item := range sortWorkoutsChronologically(workouts) {
		for _, exercise := range item.Workout.Exercises {
			point := exerciseSessionPoint{Date: item.Start, WorkoutID: item.Workout.ID}
			var rpeTotal float64
			var rpeCount int
			for _, set := range exercise.Sets {
				if set.Type == "warmup" {
					continue
				}
				if set.WeightKG != nil && *set.WeightKG > point.WeightKG {
					point.WeightKG = *set.WeightKG
				}
				if set.WeightKG != nil && set.Reps != nil {
					point.VolumeKG += *set.WeightKG * float64(*set.Reps)
					point.Reps += *set.Reps
				}
				if set.RPE != nil {
					rpeTotal += *set.RPE
					rpeCount++
				}
			}
			if rpeCount > 0 {
				point.HasRPE = true
				point.AverageRPE = rpeTotal / float64(rpeCount)
			}
			if point.WeightKG == 0 && !point.HasRPE && point.VolumeKG == 0 {
				continue
			}
			sessions[exercise.Title] = append(sessions[exercise.Title], point)
		}
	}
	return sessions
}

func buildPlateauReport(workouts []api.Workout, threshold int, includeAll bool) plateauReport {
	report := plateauReport{Threshold: threshold}
	for exercise, sessions := range buildExerciseSessions(workouts) {
		weighted := make([]exerciseSessionPoint, 0, len(sessions))
		for _, session := range sessions {
			if session.WeightKG > 0 {
				weighted = append(weighted, session)
			}
		}
		sessions = weighted
		if len(sessions) < threshold {
			continue
		}
		best := 0.0
		plateauStartIdx := -1
		for i, session := range sessions {
			if session.WeightKG > best {
				best = session.WeightKG
				plateauStartIdx = i
			}
		}
		if plateauStartIdx < 0 {
			continue
		}
		stalled := len(sessions) - plateauStartIdx
		entry := plateauEntry{
			Exercise:          exercise,
			CurrentWeightKG:   sessions[len(sessions)-1].WeightKG,
			StalledSince:      sessions[plateauStartIdx].Date,
			SessionsWithoutPR: stalled,
			Plateaued:         stalled >= threshold,
		}
		if includeAll || entry.Plateaued {
			report.Exercises = append(report.Exercises, entry)
		}
	}
	sort.Slice(report.Exercises, func(i, j int) bool {
		if report.Exercises[i].SessionsWithoutPR == report.Exercises[j].SessionsWithoutPR {
			return report.Exercises[i].Exercise < report.Exercises[j].Exercise
		}
		return report.Exercises[i].SessionsWithoutPR > report.Exercises[j].SessionsWithoutPR
	})
	return report
}

func buildSupersetPairs(workouts []api.Workout) []supersetPair {
	tally := map[string]*supersetPair{}
	for _, item := range sortWorkoutsChronologically(workouts) {
		groups := map[int][]string{}
		for _, exercise := range item.Workout.Exercises {
			var id *int
			if exercise.SupersetID != nil {
				id = exercise.SupersetID
			} else {
				id = exercise.SupersetsID
			}
			if id == nil {
				continue
			}
			groups[*id] = append(groups[*id], exercise.Title)
		}
		for _, titles := range groups {
			if len(titles) < 2 {
				continue
			}
			sort.Strings(titles)
			pairKey := strings.Join(titles, " + ")
			entry := tally[pairKey]
			if entry == nil {
				entry = &supersetPair{Pair: pairKey}
				tally[pairKey] = entry
			}
			entry.Count++
			entry.LastUsed = item.Start
		}
	}
	result := make([]supersetPair, 0, len(tally))
	for _, entry := range tally {
		result = append(result, *entry)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].LastUsed.After(result[j].LastUsed)
		}
		return result[i].Count > result[j].Count
	})
	return result
}

func buildFatigueReport(workouts []api.Workout) fatigueReport {
	report := fatigueReport{}
	for exercise, sessions := range buildExerciseSessions(workouts) {
		filtered := make([]exerciseSessionPoint, 0, len(sessions))
		for _, session := range sessions {
			if session.HasRPE {
				filtered = append(filtered, session)
			}
		}
		if len(filtered) == 0 {
			continue
		}
		window := filtered
		if len(window) > 4 {
			window = window[len(window)-4:]
		}
		entry := fatigueEntry{Exercise: exercise}
		for _, point := range window {
			entry.RPETrend = append(entry.RPETrend, round1(point.AverageRPE))
			entry.WeightTrendKG = append(entry.WeightTrendKG, point.WeightKG)
		}
		if len(window) >= 3 && window[len(window)-1].AverageRPE-window[0].AverageRPE >= 1 && window[len(window)-1].WeightKG <= window[0].WeightKG {
			entry.Signal = "FATIGUE"
			report.Signals++
		} else {
			entry.Signal = "OK"
		}
		report.Exercises = append(report.Exercises, entry)
	}
	sort.Slice(report.Exercises, func(i, j int) bool { return report.Exercises[i].Exercise < report.Exercises[j].Exercise })
	return report
}

func buildSplitReport(workouts []api.Workout, catalog map[string]api.ExerciseTemplate, weeks int, now time.Time) splitReport {
	since := now.AddDate(0, 0, -(weeks * 7))
	filtered := filterWorkoutsSince(workouts, since)
	report := splitReport{Weeks: weeks, Total: len(filtered)}
	types := map[string]*splitEntry{}
	for _, item := range sortWorkoutsChronologically(filtered) {
		splitType := categorizeWorkoutSplit(item.Workout, catalog)
		entry := types[splitType]
		if entry == nil {
			entry = &splitEntry{Type: splitType}
			types[splitType] = entry
		}
		entry.Count++
		entry.Days = append(entry.Days, item.Start.Format("Mon"))
	}
	for _, entry := range types {
		if report.Total > 0 {
			entry.Percentage = float64(entry.Count) / float64(report.Total)
		}
		report.Entries = append(report.Entries, *entry)
	}
	sort.Slice(report.Entries, func(i, j int) bool {
		if report.Entries[i].Count == report.Entries[j].Count {
			return report.Entries[i].Type < report.Entries[j].Type
		}
		return report.Entries[i].Count > report.Entries[j].Count
	})
	counts := map[string]int{}
	for _, entry := range report.Entries {
		counts[entry.Type] = entry.Count
	}
	maxMajor := maxInt(counts["Push"], maxInt(counts["Pull"], counts["Legs"]))
	if maxMajor > 0 && counts["Legs"] > 0 && counts["Legs"] < maxMajor {
		report.Warning = fmt.Sprintf("Legs slightly under-represented (%d%% vs %d%% top split). Consider adding a leg session.", pct(counts["Legs"], report.Total), pct(maxMajor, report.Total))
	}
	return report
}

func categorizeWorkoutSplit(workout api.Workout, catalog map[string]api.ExerciseTemplate) string {
	buckets := map[string]int{"push": 0, "pull": 0, "legs": 0, "other": 0}
	total := 0
	for _, exercise := range workout.Exercises {
		muscles := exerciseMuscles(exercise, catalog)
		if len(muscles) == 0 {
			continue
		}
		bucket := muscleBucket(muscles[0])
		buckets[bucket]++
		total++
	}
	if total == 0 {
		return "Full"
	}
	if float64(buckets["push"])/float64(total) > 0.5 {
		return "Push"
	}
	if float64(buckets["pull"])/float64(total) > 0.5 {
		return "Pull"
	}
	if float64(buckets["legs"])/float64(total) > 0.5 {
		return "Legs"
	}
	if buckets["legs"] >= total/2 && buckets["push"]+buckets["pull"] <= total/2 {
		return "Lower"
	}
	if buckets["push"]+buckets["pull"] >= total/2 && buckets["legs"] <= total/3 {
		return "Upper"
	}
	return "Full"
}

func buildRecordEntries(workouts []api.Workout) []recordEntry {
	byExercise := map[string][]exerciseSessionPoint{}
	for exercise, sessions := range buildExerciseSessions(workouts) {
		if len(sessions) > 0 {
			byExercise[exercise] = sessions
		}
	}
	entries := make([]recordEntry, 0, len(byExercise))
	for exercise, sessions := range byExercise {
		allTime := exerciseSessionPoint{}
		current := sessions[len(sessions)-1]
		for _, session := range sessions {
			if session.WeightKG > allTime.WeightKG || (session.WeightKG == allTime.WeightKG && session.Reps > allTime.Reps) {
				allTime = session
			}
		}
		if allTime.WeightKG == 0 && current.WeightKG == 0 {
			continue
		}
		entries = append(entries, recordEntry{
			Exercise:         exercise,
			AllTimeWeightKG:  allTime.WeightKG,
			AllTimeReps:      allTime.Reps,
			AllTimeDate:      allTime.Date,
			CurrentWeightKG:  current.WeightKG,
			CurrentReps:      current.Reps,
			CurrentDate:      current.Date,
			DeltaWeightKG:    current.WeightKG - allTime.WeightKG,
			EstimatedOneRMKG: estimateOneRM(current.WeightKG, current.Reps),
			LifetimeVolumeKG: sumVolume(sessions),
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].LifetimeVolumeKG > entries[j].LifetimeVolumeKG })
	return entries
}

func sumVolume(sessions []exerciseSessionPoint) float64 {
	total := 0.0
	for _, session := range sessions {
		total += session.VolumeKG
	}
	return total
}

func estimateOneRM(weight float64, reps int) float64 {
	if weight <= 0 || reps <= 0 {
		return 0
	}
	return weight * (1 + float64(reps)/30)
}

func buildRestEntries(workouts []api.Workout, limit int) []restEntry {
	items := sortWorkoutsChronologically(workouts)
	if len(items) > limit {
		items = items[len(items)-limit:]
	}
	entries := make([]restEntry, 0, len(items))
	for _, item := range items {
		setCount := countWorkoutSets(item.Workout)
		duration := workoutDurationMinutes(item.Workout)
		if setCount == 0 || duration <= 0 {
			continue
		}
		entries = append(entries, restEntry{
			Workout:           item.Workout.Title,
			Date:              item.Start,
			DurationMinutes:   duration,
			SetCount:          setCount,
			VolumePerMinuteKG: workoutVolume(item.Workout) / duration,
			AverageSetMinutes: duration / float64(setCount),
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Date.After(entries[j].Date) })
	return entries
}

func countWorkoutSets(workout api.Workout) int {
	total := 0
	for _, exercise := range workout.Exercises {
		total += len(exercise.Sets)
	}
	return total
}

func formatDaysAgo(days int) string {
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}

func formatTrend(values []float64, weight bool) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if weight {
			parts = append(parts, formatWeight(value))
		} else {
			parts = append(parts, fmt.Sprintf("%.1f", value))
		}
	}
	return strings.Join(parts, "→")
}

func recordDeltaLabel(delta float64) string {
	if math.Abs(delta) < 0.001 {
		return "="
	}
	if delta > 0 {
		return "+" + formatWeight(delta)
	}
	return formatWeight(delta)
}

func countPlateaus(entries []plateauEntry) int {
	total := 0
	for _, entry := range entries {
		if entry.Plateaued {
			total++
		}
	}
	return total
}

func pct(value, total int) int {
	if total == 0 {
		return 0
	}
	return int(math.Round(float64(value) * 100 / float64(total)))
}

func efficiencyBar(value, max float64) string {
	if max <= 0 {
		return "█"
	}
	width := int(math.Round((value / max) * 14))
	if width < 1 {
		width = 1
	}
	return strings.Repeat("█", width)
}

func round1(value float64) float64 {
	return math.Round(value*10) / 10
}

func defaultWhoopPath() string {
	return filepath.Join("~", ".openclaw", "workspace", "skills", "whoop-tracker")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
