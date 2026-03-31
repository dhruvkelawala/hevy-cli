package cmd

import (
	"fmt"
	"os"

	"github.com/dhruvkelawala/go-hevy/internal/api"
	"github.com/spf13/cobra"
)

var exportFormat string
var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export workouts in a machine-readable format",
	RunE: func(cmd *cobra.Command, args []string) error {
		if exportFormat != "csv" {
			return fmt.Errorf("unsupported format %q", exportFormat)
		}
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 0, true)
		if err != nil {
			return err
		}
		detailed := make([]api.Workout, 0, len(workouts))
		for _, workout := range workouts {
			detail, err := client.GetWorkout(contextForCommand(cmd), workout.ID)
			if err != nil {
				return err
			}
			detailed = append(detailed, *detail)
		}
		csvText, err := encodeCSV(workoutCSVRows(detailed))
		if err != nil {
			return err
		}
		if exportOutput != "" {
			return os.WriteFile(exportOutput, []byte(csvText), 0o644)
		}
		_, err = fmt.Fprint(os.Stdout, csvText)
		return err
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "csv", "Export format")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Write export to a file instead of stdout")
}
