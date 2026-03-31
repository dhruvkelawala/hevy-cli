package cmd

import (
	"fmt"
	"os"

	appconfig "github.com/dhruvkelawala/hevy-cli/internal/config"
	"github.com/dhruvkelawala/hevy-cli/internal/output"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or update configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch app.outputMode {
		case outputJSON:
			return output.PrintJSON(os.Stdout, map[string]any{
				"config_path":    configRows()[0][1],
				"api_key":        appconfig.Redact(app.config.EffectiveAPIKey()),
				"unit":           app.config.Unit,
				"default_limit":  app.config.DefaultLimit,
				"api_key_source": configSource(),
			})
		case outputCompact:
			return output.PrintCompact(os.Stdout, []string{
				fmt.Sprintf("config_path=%s", configRows()[0][1]),
				fmt.Sprintf("api_key=%s", appconfig.Redact(app.config.EffectiveAPIKey())),
				fmt.Sprintf("unit=%s", output.ValueOrDash(app.config.Unit)),
				fmt.Sprintf("default_limit=%d", app.config.DefaultLimit),
				fmt.Sprintf("api_key_source=%s", configSource()),
			})
		default:
			output.PrintKeyValueTable(os.Stdout, configRows())
			return nil
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a stored config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := app.config
		if cfg == nil {
			cfg = &appconfig.Config{}
		}
		switch args[0] {
		case "key", "api_key":
			cfg.APIKey = args[1]
		case "unit":
			if args[1] != "kg" && args[1] != "lbs" {
				return fmt.Errorf("unit must be kg or lbs")
			}
			cfg.Unit = args[1]
		case "default_limit":
			var limit int
			if _, err := fmt.Sscanf(args[1], "%d", &limit); err != nil || limit <= 0 {
				return fmt.Errorf("default_limit must be a positive integer")
			}
			cfg.DefaultLimit = limit
		default:
			return fmt.Errorf("unsupported config key %q", args[0])
		}
		if err := appconfig.Save(cfg); err != nil {
			return err
		}
		app.config = cfg
		printLine("Updated config key %q", args[0])
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
}
