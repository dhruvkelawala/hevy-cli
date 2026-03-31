package cmd

import (
	"fmt"
	"os"

	"github.com/dhruvkelawala/hevy-cli/internal/api"
	"github.com/dhruvkelawala/hevy-cli/internal/output"
	"github.com/spf13/cobra"
)

var historyStartDate string
var historyEndDate string

var historyCmd = &cobra.Command{
	Use:   "history <exercise-template-id>",
	Short: "Show history for an exercise template",
	Long:  "Show logged set history for an exercise template ID from `hevy exercises` or a workout's `exercise_template_id`.",
	Example: "  hevy history 7EB3F7C3\n" +
		"  hevy history 7EB3F7C3 --start-date 2026-01-01T00:00:00Z",
	Args: requireSingleIdentifierArg("exercise-template-id"),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		resp, err := client.GetExerciseHistory(contextForCommand(cmd), args[0], historyStartDate, historyEndDate)
		if err != nil {
			return err
		}
		return renderExerciseHistory(args[0], resp, historyStartDate != "" || historyEndDate != "")
	},
}

func renderExerciseHistory(exerciseTemplateID string, resp *api.ExerciseHistoryResponse, filtered bool) error {
	if app.outputMode == outputJSON {
		return output.PrintJSON(os.Stdout, resp)
	}
	if resp == nil || len(resp.ExerciseHistory) == 0 {
		return output.PrintCompact(os.Stdout, historyEmptyStateLines(exerciseTemplateID, filtered))
	}
	if app.outputMode == outputCompact {
		lines := make([]string, 0, len(resp.ExerciseHistory))
		for _, entry := range resp.ExerciseHistory {
			lines = append(lines, fmt.Sprintf("%s | %s | weight=%s | reps=%s", entry.WorkoutID, entry.WorkoutTitle, formatFloatPtr(entry.WeightKG), formatIntPtr(entry.Reps)))
		}
		return output.PrintCompact(os.Stdout, lines)
	}
	rows := make([][]string, 0, len(resp.ExerciseHistory))
	for _, entry := range resp.ExerciseHistory {
		rows = append(rows, []string{entry.WorkoutID, entry.WorkoutTitle, formatTimestamp(entry.WorkoutStartTime), entry.SetType, formatFloatPtr(entry.WeightKG), formatIntPtr(entry.Reps), formatFloatPtr(entry.RPE)})
	}
	output.PrintTable(os.Stdout, []string{"Workout ID", "Workout", "Start", "Set Type", "Weight KG", "Reps", "RPE"}, rows)
	return nil
}

func init() {
	historyCmd.Flags().StringVar(&historyStartDate, "start-date", "", "Filter history from ISO-8601 timestamp")
	historyCmd.Flags().StringVar(&historyEndDate, "end-date", "", "Filter history to ISO-8601 timestamp")
}
