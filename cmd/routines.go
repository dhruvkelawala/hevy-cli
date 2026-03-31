package cmd

import (
	"fmt"
	"os"

	"github.com/dhruvkelawala/hevy-cli/internal/api"
	"github.com/dhruvkelawala/hevy-cli/internal/output"
	"github.com/spf13/cobra"
)

var routinesCmd = &cobra.Command{
	Use:   "routines",
	Short: "List routines",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requirePositivePagination(app.page, app.pageSize, 10); err != nil {
			return err
		}
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		resp, err := client.ListRoutines(contextForCommand(cmd), app.page, app.pageSize)
		if err != nil {
			return err
		}
		switch app.outputMode {
		case outputJSON:
			return output.PrintJSON(os.Stdout, resp)
		case outputCompact:
			lines := make([]string, 0, len(resp.Routines))
			for _, routine := range resp.Routines {
				lines = append(lines, fmt.Sprintf("%s | %s | exercises=%d", routine.ID, routine.Title, len(routine.Exercises)))
			}
			return output.PrintCompact(os.Stdout, lines)
		default:
			rows := make([][]string, 0, len(resp.Routines))
			for _, routine := range resp.Routines {
				rows = append(rows, []string{routine.ID, routine.Title, formatTimestamp(routine.CreatedAt), formatTimestamp(routine.UpdatedAt)})
			}
			output.PrintTable(os.Stdout, []string{"ID", "Title", "Created", "Updated"}, rows)
			return nil
		}
	},
}

var routineCmd = &cobra.Command{
	Use:   "routine <routine-id>",
	Short: "Get a routine",
	Args:  requireSingleIdentifierArg("routine-id"),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		routine, err := client.GetRoutine(contextForCommand(cmd), args[0])
		if err != nil {
			return err
		}
		return renderRoutine(routine)
	},
}

func renderRoutine(routine *api.Routine) error {
	if routine == nil {
		return fmt.Errorf("routine not found")
	}
	if app.outputMode == outputJSON {
		return output.PrintJSON(os.Stdout, routine)
	}
	compact := []string{fmt.Sprintf("%s | %s | exercises=%d", routine.ID, routine.Title, len(routine.Exercises))}
	if app.outputMode == outputCompact {
		return output.PrintCompact(os.Stdout, compact)
	}

	output.PrintKeyValueTable(os.Stdout, [][2]string{{"id", routine.ID}, {"title", routine.Title}, {"notes", output.ValueOrDash(routine.Notes)}, {"created_at", formatTimestamp(routine.CreatedAt)}, {"updated_at", formatTimestamp(routine.UpdatedAt)}})
	if len(routine.Exercises) > 0 {
		fmt.Fprintln(os.Stdout)
		rows := make([][]string, 0, len(routine.Exercises))
		for _, exercise := range routine.Exercises {
			rows = append(rows, []string{exercise.ExerciseTemplateID, output.ValueOrDash(exercise.Title), fmt.Sprintf("%d", len(exercise.Sets)), formatIntPtr(exercise.RestSeconds), output.ValueOrDash(exercise.Notes)})
		}
		output.PrintTable(os.Stdout, []string{"Exercise ID", "Title", "Sets", "Rest S", "Notes"}, rows)
	}
	return nil
}
