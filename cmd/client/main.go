package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/adrg/xdg"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/cmd"
	"github.com/nt0xa/sonar/internal/modules/api/apiclient"
	"github.com/nt0xa/sonar/internal/templates"
)

func init() {
	validation.ErrorTag = "err"
}

func main() {
	var (
		cfg        Config
		cfgFile    string
		jsonOutput bool
	)

	c := cmd.New(
		nil,
		cmd.AllowFileAccess(true),
		cmd.PreExec(func(acts *actions.Actions, root *cobra.Command) {
			root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
				// We have to find the flag value ourselves because flags are not parsed on completion.
				if err := initConfig(findFlagValue("config", os.Args), &cfg); err != nil {
					return err
				}

				if err := initActions(acts, cfg); err != nil {
					return err
				}

				return nil
			}

			// Flags, commands...
			root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
			jsonFlag(root, &jsonOutput)
			contextCmd(root, &cfg)
		}),
	)

	stdout, stderr, err := c.Exec(context.Background(), os.Args[1:], func(res actions.Result) error {
		if jsonOutput {
			return json.NewEncoder(os.Stdout).Encode(res)
		}
		tmpl := templates.New(cfg.Server().BaseURL().Hostname(),
			templates.Default(
				templates.HTMLEscape(false),
				templates.Markup(templates.Bold("<bold>", "</>"))),
		)
		s, err := tmpl.RenderResult(res)
		if err != nil {
			return err
		}
		color.Fprint(os.Stdout, s)
		return nil
	})
	cobra.CheckErr(err)

	if stdout != "" {
		fmt.Fprint(os.Stdout, stdout)
	}

	if stderr != "" {
		fmt.Fprint(os.Stderr, stderr)
	}
}

func jsonFlag(root *cobra.Command, jsonOutput *bool) {
	for _, cmd := range root.Commands() {
		if cmd.HasSubCommands() {
			jsonFlag(cmd, jsonOutput)
		}

		if cmd.Name() == "help" ||
			cmd.Name() == "completion" ||
			strings.HasPrefix(cmd.Name(), "_") {
			continue
		}

		cmd.Flags().BoolVar(jsonOutput, "json", false, "JSON output")
	}
}

func contextCmd(root *cobra.Command, cfg *Config) {
	var server string

	cmd := &cobra.Command{
		Use:   "ctx",
		Short: "Change current context parameters",
		RunE: func(cmd *cobra.Command, args []string) error {

			// Save current config
			if err := viper.WriteConfig(); err != nil {
				return fmt.Errorf("fail to save config: %w", err)
			}

			// Print values from current context.
			fields := reflect.VisibleFields(reflect.TypeOf(cfg.Context))
			v := reflect.ValueOf(cfg.Context)
			for _, field := range fields {
				color.Bold.Print(field.Tag.Get("mapstructure") + ": ")
				color.Println(v.FieldByName(field.Name))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&server, "server", "s", "", "Server name from list of servers")
	viper.BindPFlag("context.server", cmd.Flags().Lookup("server"))

	cmd.RegisterFlagCompletionFunc("server", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return maps.Keys(cfg.Servers), cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(cmd)
}

func initConfig(cfgFile string, cfg *Config) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		configFilePath, err := xdg.ConfigFile("sonar/config.toml")
		if err != nil {
			cobra.CheckErr(err)
		}
		viper.SetConfigFile(configFilePath)
	}

	viper.SetEnvPrefix("sonar")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return err
	}

	if err := cfg.ValidateWithContext(context.Background()); err != nil {
		return err
	}

	return nil
}

func initActions(acts *actions.Actions, cfg Config) error {
	srv := cfg.Server()
	if srv == nil {
		return errors.New("server must be set")
	}
	*acts = apiclient.New(srv.URL, srv.Token, srv.Insecure, srv.Proxy)
	return nil
}

func findFlagValue(f string, args []string) string {
	for i := 1; i < len(args); i++ {
		if args[i-1] == "--"+f {
			return args[i]
		}
	}
	return ""
}
