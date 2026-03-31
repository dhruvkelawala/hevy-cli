package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/dhruvkelawala/hevy-cli/internal/api"
	"github.com/dhruvkelawala/hevy-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	planDays             int
	consistencyMonths    int
	consistencyTarget    int
	plateauThreshold     int
	plateauAll           bool
	splitWeeks           int
	recordsTop           int
	restLimit            int
	readinessNoWhoop     bool
	readinessHistoryDays int
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Suggest your next workout",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requirePositiveInt("days", planDays); err != nil {
			return err
		}
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		workouts, err = ensureWorkoutDetails(client, workouts)
		if err != nil {
			return err
		}
		exercises, err := fetchAllExercises(client)
		if err != nil {
			return err
		}
		routines, err := fetchAllRoutines(client)
		if err != nil {
			return err
		}

		now := time.Now()
		filtered := filterWorkoutsSince(workouts, now.AddDate(0, 0, -planDays))
		report := recommendRoutine(filtered, routines, exerciseCatalogMap(exercises), planDays, now)
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, report)
		}
		if len(report.LastTrained) == 0 {
			printLine("No workouts found in the last %d days.", planDays)
		} else {
			printLine("Last trained:")
			for _, entry := range report.LastTrained {
				printLine("  %-14s %s (%s)", strings.ReplaceAll(entry.Muscle, "_", " "), formatDaysAgo(entry.DaysAgo), entry.Weekday)
			}
		}
		if report.SuggestedRoutine != "" {
			printLine("")
			printLine("Suggested next: %s", report.SuggestedRoutine)
		}
		printLine("Reason: %s", report.Reason)
		return nil
	},
}

var consistencyCmd = &cobra.Command{
	Use:   "consistency",
	Short: "Show training consistency over time",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requirePositiveInt("months", consistencyMonths); err != nil {
			return err
		}
		if err := requirePositiveInt("target", consistencyTarget); err != nil {
			return err
		}
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		workouts, err = ensureWorkoutDetails(client, workouts)
		if err != nil {
			return err
		}
		report := buildConsistencyReport(workouts, consistencyMonths, consistencyTarget, time.Now())
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, report)
		}
		printLine("Monthly Consistency (last %d months)", consistencyMonths)
		rows := make([][]string, 0, len(report.Monthly))
		for i := len(report.Monthly) - 1; i >= 0; i-- {
			entry := report.Monthly[i]
			rows = append(rows, []string{
				entry.Month,
				fmt.Sprintf("%d", entry.Sessions),
				fmt.Sprintf("%d", entry.Target),
				fmt.Sprintf("%d%%", int(entry.HitRate*100+0.5)),
				formatWeight(entry.TotalVolumeKG) + app.weightUnit,
				fmt.Sprintf("%d days", entry.LongestGapDays),
			})
		}
		output.PrintTable(os.Stdout, []string{"Month", "Sessions", "Target", "Hit%", "Volume", "Longest Gap"}, rows)
		printLine("")
		printLine("Overall: %d/%d sessions (%d%%)", report.OverallSessions, report.OverallTarget, int(report.OverallHitRate*100+0.5))
		printLine("Current streak: %d weeks without missing", report.CurrentTargetStreak)
		return nil
	},
}

var plateauCmd = &cobra.Command{
	Use:   "plateau",
	Short: "Detect exercises that have stalled",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requirePositiveInt("threshold", plateauThreshold); err != nil {
			return err
		}
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		workouts, err = ensureWorkoutDetails(client, workouts)
		if err != nil {
			return err
		}
		report := buildPlateauReport(workouts, plateauThreshold, plateauAll)
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, report)
		}
		if len(report.Exercises) == 0 {
			printLine("No plateaued exercises found.")
			return nil
		}
		rows := make([][]string, 0, len(report.Exercises))
		for _, entry := range report.Exercises {
			rows = append(rows, []string{
				entry.Exercise,
				formatWeight(entry.CurrentWeightKG) + app.weightUnit,
				entry.StalledSince.Format("Jan 2"),
				fmt.Sprintf("%d sessions", entry.SessionsWithoutPR),
			})
		}
		output.PrintTable(os.Stdout, []string{"Exercise", weightHeader(), "Stalled Since", "Sessions"}, rows)
		if !plateauAll {
			printLine("")
			printLine("%d exercises plateaued. Consider: deload, rep scheme change, or micro-loading.", countPlateaus(report.Exercises))
		}
		return nil
	},
}

var supersetsCmd = &cobra.Command{
	Use:   "supersets",
	Short: "Show your most-used superset pairings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		workouts, err = ensureWorkoutDetails(client, workouts)
		if err != nil {
			return err
		}
		filtered := filterWorkoutsSince(workouts, time.Now().AddDate(0, 0, -30))
		pairs := buildSupersetPairs(filtered)
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, map[string]any{"days": 30, "pairs": pairs})
		}
		if len(pairs) == 0 {
			printLine("No supersets found in recent workouts.")
			return nil
		}
		rows := make([][]string, 0, len(pairs))
		for _, pair := range pairs {
			rows = append(rows, []string{pair.Pair, fmt.Sprintf("%d", pair.Count), pair.LastUsed.Format("Jan 2")})
		}
		output.PrintTable(os.Stdout, []string{"Superset Pair", "Count", "Last Used"}, rows)
		return nil
	},
}

var fatigueCmd = &cobra.Command{
	Use:   "fatigue",
	Short: "Analyze RPE trends for fatigue signals",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		workouts, err = ensureWorkoutDetails(client, workouts)
		if err != nil {
			return err
		}
		report := buildFatigueReport(workouts)
		if len(report.Exercises) == 0 {
			report.Message = "No RPE data found. Log RPE in Hevy to enable fatigue tracking."
		}
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, report)
		}
		if len(report.Exercises) == 0 {
			printLine(report.Message)
			return nil
		}
		rows := make([][]string, 0, len(report.Exercises))
		for _, entry := range report.Exercises {
			signal := "OK"
			if entry.Signal == "FATIGUE" {
				signal = "⚠️ FATIGUE"
			}
			rows = append(rows, []string{entry.Exercise, formatTrend(entry.RPETrend, false), formatTrend(entry.WeightTrendKG, true), signal})
		}
		output.PrintTable(os.Stdout, []string{"Exercise", "RPE Trend", "Weight Trend", "Signal"}, rows)
		printLine("")
		if report.Signals > 0 {
			printLine("⚠️ %d exercise(s) showing fatigue signal. Consider a deload week.", report.Signals)
		} else {
			printLine("No clear fatigue signals detected.")
		}
		return nil
	},
}

var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Analyze your actual training split",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requirePositiveInt("weeks", splitWeeks); err != nil {
			return err
		}
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		workouts, err = ensureWorkoutDetails(client, workouts)
		if err != nil {
			return err
		}
		exercises, err := fetchAllExercises(client)
		if err != nil {
			return err
		}
		report := buildSplitReport(workouts, exerciseCatalogMap(exercises), splitWeeks, time.Now())
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, report)
		}
		printLine("Training Split (last %d weeks)", splitWeeks)
		rows := make([][]string, 0, len(report.Entries))
		for _, entry := range report.Entries {
			rows = append(rows, []string{entry.Type, fmt.Sprintf("%d", entry.Count), fmt.Sprintf("%d%%", int(entry.Percentage*100+0.5)), strings.Join(entry.Days, ", ")})
		}
		output.PrintTable(os.Stdout, []string{"Type", "Count", "Percentage", "Days"}, rows)
		if report.Warning != "" {
			printLine("")
			printLine("⚠️ %s", report.Warning)
		}
		return nil
	},
}

var recordsCmd = &cobra.Command{
	Use:   "records",
	Short: "Compare all-time records against current bests",
	RunE: func(cmd *cobra.Command, args []string) error {
		if recordsTop < 0 {
			return fmt.Errorf("top must be 0 or greater")
		}
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		workouts, err = ensureWorkoutDetails(client, workouts)
		if err != nil {
			return err
		}
		entries := buildRecordEntries(workouts)
		if recordsTop > 0 && len(entries) > recordsTop {
			entries = entries[:recordsTop]
		}
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, map[string]any{"records": entries})
		}
		if len(entries) == 0 {
			printLine("No weighted records found.")
			return nil
		}
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{
				entry.Exercise,
				fmt.Sprintf("%s×%d (%s)", formatWeight(entry.AllTimeWeightKG), entry.AllTimeReps, entry.AllTimeDate.Format("Jan")),
				fmt.Sprintf("%s×%d (%s)", formatWeight(entry.CurrentWeightKG), entry.CurrentReps, entry.CurrentDate.Format("Jan")),
				recordDeltaLabel(entry.DeltaWeightKG),
				formatWeight(entry.EstimatedOneRMKG) + app.weightUnit,
			})
		}
		output.PrintTable(os.Stdout, []string{"Exercise", "All-Time Best", "Current Best", "Δ", "Est 1RM"}, rows)
		return nil
	},
}

var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "Estimate workout time efficiency",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requirePositiveInt("limit", restLimit); err != nil {
			return err
		}
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, restLimit, false)
		if err != nil {
			return err
		}
		workouts, err = ensureWorkoutDetails(client, workouts)
		if err != nil {
			return err
		}
		entries := buildRestEntries(workouts, restLimit)
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, map[string]any{"sessions": entries})
		}
		if len(entries) == 0 {
			printLine("No workouts with duration and set data found.")
			return nil
		}
		maxVPM := 0.0
		for _, entry := range entries {
			if entry.VolumePerMinuteKG > maxVPM {
				maxVPM = entry.VolumePerMinuteKG
			}
		}
		printLine("Time Efficiency (last %d sessions)", restLimit)
		rows := make([][]string, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, []string{
				entry.Workout,
				entry.Date.Format("Jan 2"),
				fmt.Sprintf("%.0fmin", entry.DurationMinutes),
				fmt.Sprintf("%d", entry.SetCount),
				fmt.Sprintf("%s%s", formatWeight(entry.VolumePerMinuteKG), app.weightUnit),
				efficiencyBar(entry.VolumePerMinuteKG, maxVPM),
			})
		}
		output.PrintTable(os.Stdout, []string{"Workout", "Date", "Duration", "Sets", "Vol/Min", "Efficiency"}, rows)
		return nil
	},
}

var readinessCmd = &cobra.Command{
	Use:   "readiness",
	Short: "Combine recovery and training readiness",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requirePositiveInt("days", readinessHistoryDays); err != nil {
			return err
		}
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		workouts, err = ensureWorkoutDetails(client, workouts)
		if err != nil {
			return err
		}
		exercises, err := fetchAllExercises(client)
		if err != nil {
			return err
		}
		routines, err := fetchAllRoutines(client)
		if err != nil {
			return err
		}

		now := time.Now()
		catalog := exerciseCatalogMap(exercises)
		report := readinessReport{}
		plan := recommendRoutine(filterWorkoutsSince(workouts, now.AddDate(0, 0, -14)), routines, catalog, 14, now)
		if plan.SuggestedRoutine != "" || len(plan.LastTrained) > 0 || plan.Reason != "" {
			report.Plan = &plan
		}
		if lastWorkout, lastDate, ok := latestWorkout(workouts); ok {
			report.LastWorkoutTitle = lastWorkout.Title
			report.LastWorkoutDate = &lastDate
			report.AvoidMuscles = workoutMuscles(lastWorkout, catalog)
		}

		if readinessNoWhoop {
			report.Status, report.Advice = readinessWithoutWhoop(report.AvoidMuscles, report.LastWorkoutDate, now)
		} else {
			todayResp, whoopErr := fetchWhoopRecovery("--today", "--json")
			if whoopErr != nil {
				report.WhoopMessage = "WHOOP not configured. Run: hevy config set whoop_path <path-to-whoop-skill>"
				report.Status, report.Advice = readinessWithoutWhoop(report.AvoidMuscles, report.LastWorkoutDate, now)
			} else {
				report.Whoop = parseWhoopSnapshot(todayResp)
				if report.Whoop != nil {
					report.Status, report.Advice = whoopStatus(report.Whoop.RecoveryScore)
				} else {
					report.WhoopMessage = "WHOOP recovery data not available for today."
					report.Status, report.Advice = readinessWithoutWhoop(report.AvoidMuscles, report.LastWorkoutDate, now)
				}
				historyResp, err := fetchWhoopRecovery("--days", fmt.Sprintf("%d", readinessHistoryDays), "--json")
				if err == nil {
					report.RecoveryHistory = parseWhoopHistory(historyResp)
				}
			}
		}

		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, report)
		}
		if report.Whoop != nil {
			indicator := map[string]string{"GREEN": "🟢", "YELLOW": "🟡", "RED": "🔴"}[report.Status]
			printLine("%s WHOOP Recovery: %.0f%%  |  HRV: %.0fms  |  RHR: %.0fbpm", indicator, report.Whoop.RecoveryScore, report.Whoop.HRVRMSSD, report.Whoop.RestingHR)
			printLine("")
		} else if report.WhoopMessage != "" {
			printLine(report.WhoopMessage)
			printLine("")
		}
		printLine("Status: %s — %s", report.Status, strings.ToLower(report.Advice))
		if report.LastWorkoutTitle != "" && report.LastWorkoutDate != nil {
			printLine("Last trained: %s (%s)", report.LastWorkoutTitle, report.LastWorkoutDate.Format("Jan 2"))
		}
		if len(report.AvoidMuscles) > 0 {
			printLine("Avoid today: %s", strings.Join(report.AvoidMuscles, ", "))
		}
		if report.Plan != nil && report.Plan.SuggestedRoutine != "" {
			printLine("Suggested: %s", report.Plan.SuggestedRoutine)
		}
		if len(report.RecoveryHistory) > 0 {
			printLine("")
			printLine("Recovery history (last %d days):", readinessHistoryDays)
			for _, point := range report.RecoveryHistory {
				printLine("%s  %3.0f%%  %s", point.Day, point.RecoveryScore, recoveryBar(point.RecoveryScore))
			}
		}
		return nil
	},
}

func latestWorkout(workouts []api.Workout) (api.Workout, time.Time, bool) {
	items := sortWorkoutsChronologically(workouts)
	if len(items) == 0 {
		return api.Workout{}, time.Time{}, false
	}
	item := items[len(items)-1]
	return item.Workout, item.Start, true
}

func workoutMuscles(workout api.Workout, catalog map[string]api.ExerciseTemplate) []string {
	muscles := []string{}
	for _, exercise := range workout.Exercises {
		for _, muscle := range exerciseMuscles(exercise, catalog) {
			if !containsString(muscles, muscle) {
				muscles = append(muscles, muscle)
			}
		}
	}
	sort.Strings(muscles)
	return muscles
}

func readinessWithoutWhoop(avoid []string, lastWorkoutDate *time.Time, now time.Time) (string, string) {
	if lastWorkoutDate != nil && now.Sub(*lastWorkoutDate).Hours() < 36 {
		if len(avoid) > 0 {
			return "CAUTION", fmt.Sprintf("Recent session on %s. Keep effort moderate and avoid %s.", lastWorkoutDate.Format("Jan 2"), strings.Join(avoid, ", "))
		}
		return "CAUTION", fmt.Sprintf("Recent session on %s. Keep effort moderate today.", lastWorkoutDate.Format("Jan 2"))
	}
	if len(avoid) > 0 {
		return "READY", fmt.Sprintf("History-only mode. Prefer muscles other than %s today.", strings.Join(avoid, ", "))
	}
	return "READY", "History-only mode. No obvious recovery conflicts found."
}

func recoveryBar(score float64) string {
	width := int(score / 10)
	if width < 1 {
		width = 1
	}
	if width > 10 {
		width = 10
	}
	return strings.Repeat("█", width)
}

func requirePositiveInt(name string, value int) error {
	if value <= 0 {
		return fmt.Errorf("%s must be greater than 0", name)
	}
	return nil
}

func init() {
	planCmd.Flags().IntVar(&planDays, "days", 14, "Lookback window in days")
	consistencyCmd.Flags().IntVar(&consistencyMonths, "months", 3, "Months to include")
	consistencyCmd.Flags().IntVar(&consistencyTarget, "target", 4, "Target sessions per week")
	plateauCmd.Flags().IntVar(&plateauThreshold, "threshold", 3, "Consecutive sessions without improvement")
	plateauCmd.Flags().BoolVar(&plateauAll, "all", false, "Show all exercises, not just plateaued ones")
	splitCmd.Flags().IntVar(&splitWeeks, "weeks", 4, "Weeks to include")
	recordsCmd.Flags().IntVar(&recordsTop, "top", 0, "Show top N exercises by lifetime volume (0 = all)")
	restCmd.Flags().IntVar(&restLimit, "limit", 10, "Number of recent sessions to inspect")
	readinessCmd.Flags().BoolVar(&readinessNoWhoop, "no-whoop", false, "Skip WHOOP integration")
	readinessCmd.Flags().IntVar(&readinessHistoryDays, "days", 7, "Recovery history days")

	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(consistencyCmd)
	rootCmd.AddCommand(plateauCmd)
	rootCmd.AddCommand(supersetsCmd)
	rootCmd.AddCommand(fatigueCmd)
	rootCmd.AddCommand(splitCmd)
	rootCmd.AddCommand(recordsCmd)
	rootCmd.AddCommand(restCmd)
	rootCmd.AddCommand(readinessCmd)
}
