package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dhruvkelawala/hevy-cli/internal/api"
	"github.com/dhruvkelawala/hevy-cli/internal/output"
	"github.com/spf13/cobra"
)

var exerciseSearch string
var exerciseMuscle string
var exerciseCustomOnly bool

var exercisesCmd = &cobra.Command{
	Use:   "exercises",
	Short: "List exercise templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requirePositivePagination(app.page, app.pageSize, 100); err != nil {
			return err
		}
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		resp := &api.PaginatedExerciseTemplates{Page: app.page, PageCount: 1}
		if needsExerciseCatalog(exerciseSearch, exerciseMuscle, exerciseCustomOnly) {
			allExercises, err := fetchAllExercises(client)
			if err != nil {
				return err
			}
			resp.ExerciseTemplates = filterExercises(allExercises, exerciseSearch, exerciseMuscle, exerciseCustomOnly)
		} else {
			pagedResp, err := client.ListExercises(contextForCommand(cmd), app.page, app.pageSize)
			if err != nil {
				return err
			}
			resp = pagedResp
		}
		switch app.outputMode {
		case outputJSON:
			return output.PrintJSON(os.Stdout, resp)
		case outputCompact:
			lines := make([]string, 0, len(resp.ExerciseTemplates))
			for _, exercise := range resp.ExerciseTemplates {
				lines = append(lines, fmt.Sprintf("%s | %s | %s", exercise.ID, exercise.Title, exercise.PrimaryMuscleGroup))
			}
			return output.PrintCompact(os.Stdout, lines)
		default:
			rows := make([][]string, 0, len(resp.ExerciseTemplates))
			for _, exercise := range resp.ExerciseTemplates {
				rows = append(rows, []string{exercise.ID, exercise.Title, exercise.Type, exercise.PrimaryMuscleGroup})
			}
			output.PrintTable(os.Stdout, []string{"ID", "Title", "Type", "Primary Muscle"}, rows)
			return nil
		}
	},
}

func init() {
	exercisesCmd.Flags().StringVar(&exerciseSearch, "search", "", "Search exercises by name")
	exercisesCmd.Flags().StringVar(&exerciseMuscle, "muscle", "", "Filter by primary or secondary muscle group")
	exercisesCmd.Flags().BoolVar(&exerciseCustomOnly, "custom", false, "Show only custom exercises")
}

var exerciseCmd = &cobra.Command{
	Use:   "exercise <exercise-id>",
	Short: "Get an exercise template",
	Args:  requireSingleIdentifierArg("exercise-id"),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := clientFromConfig()
		if err != nil {
			return err
		}
		exercise, err := client.GetExercise(contextForCommand(cmd), args[0])
		if err != nil {
			return err
		}
		return renderExercise(exercise)
	},
}

func renderExercise(exercise *api.ExerciseTemplate) error {
	if exercise == nil {
		return fmt.Errorf("exercise not found")
	}
	if app.outputMode == outputJSON {
		return output.PrintJSON(os.Stdout, exercise)
	}
	compact := []string{fmt.Sprintf("%s | %s | %s", exercise.ID, exercise.Title, exercise.PrimaryMuscleGroup)}
	if app.outputMode == outputCompact {
		return output.PrintCompact(os.Stdout, compact)
	}
	output.PrintKeyValueTable(os.Stdout, [][2]string{{"id", exercise.ID}, {"title", exercise.Title}, {"type", exercise.Type}, {"primary_muscle_group", output.ValueOrDash(exercise.PrimaryMuscleGroup)}, {"secondary_muscle_groups", output.ValueOrDash(strings.Join(exercise.SecondaryMuscleGroups, ", "))}, {"is_custom", fmt.Sprintf("%t", exercise.IsCustom)}})
	return nil
}
