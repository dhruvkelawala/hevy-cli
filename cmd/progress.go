package cmd

import (
	"fmt"
	"os"

	"github.com/dhruvkelawala/go-hevy/internal/output"
	"github.com/spf13/cobra"
)

var progressCmd = &cobra.Command{
	Use:   "progress <exercise-name>",
	Short: "Show an ASCII progression chart for an exercise",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		exercises, err := fetchAllExercises(client)
		if err != nil {
			return err
		}
		exercise := pickExerciseByName(exercises, args[0])
		if exercise == nil {
			return fmt.Errorf("no exercise found matching %q", args[0])
		}
		history, err := client.GetExerciseHistory(contextForCommand(cmd), exercise.ID, "", "")
		if err != nil {
			return err
		}
		points := buildProgressPoints(history.ExerciseHistory)
		lines := renderProgressChart(exercise.Title, points, 8)
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, map[string]any{"exercise": exercise, "points": points, "chart": lines})
		}
		return output.PrintCompact(os.Stdout, lines)
	},
}
