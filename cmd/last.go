package cmd

import "github.com/spf13/cobra"

var lastCmd = &cobra.Command{
	Use:   "last",
	Short: "Show the most recent workout in detail",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		workouts, err := fetchWorkouts(client, 1, 1, false)
		if err != nil {
			return err
		}
		if len(workouts) == 0 {
			return nil
		}
		return runWorkoutGet(cmd, workouts[0].ID)
	},
}
