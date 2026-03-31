package cmd

import (
	"fmt"
	"os"

	"github.com/dhruvkelawala/go-hevy/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Verify API access and show account status",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		profile, err := client.GetMe(contextForCommand(cmd))
		if err != nil {
			return err
		}
		count, err := client.GetWorkoutCount(contextForCommand(cmd))
		if err != nil {
			return err
		}
		status := map[string]any{"api_key_valid": true, "workout_count": count.WorkoutCount, "profile": profile}
		if app.outputMode == outputJSON {
			return output.PrintJSON(os.Stdout, status)
		}
		name := fmt.Sprintf("%v", profile["name"])
		id := fmt.Sprintf("%v", profile["id"])
		url := fmt.Sprintf("%v", profile["url"])
		rows := [][2]string{{"api_key_valid", "true"}, {"name", name}, {"id", id}, {"url", url}, {"workout_count", fmt.Sprintf("%d", count.WorkoutCount)}}
		if app.outputMode == outputCompact {
			return output.PrintCompact(os.Stdout, []string{fmt.Sprintf("api_key_valid=true name=%s id=%s workout_count=%d", name, id, count.WorkoutCount)})
		}
		output.PrintKeyValueTable(os.Stdout, rows)
		return nil
	},
}
