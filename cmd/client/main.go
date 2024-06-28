package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/adrg/xdg"
	"github.com/carapace-sh/carapace"
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

var (
	cfg        Config
	cfgFile    string
	jsonOutput bool
)

func init() {
	validation.ErrorTag = "err"
	cobra.OnInitialize(initConfig)
}

func main() {
	f, _ := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	slog.SetDefault(slog.New(slog.NewTextHandler(f, nil)))
	defer f.Close()

	c := cmd.New(
		nil,
		cmd.AllowFileAccess(true),
		cmd.PreExec(func(root *cobra.Command) {
			addConfigFlag(root)
			addJSONFlag(root)
			addContextCommand(root)

			carapace.Gen(root).Standalone()
		}),
		cmd.InitActions(func() (actions.Actions, error) {
			srv := cfg.Server()
			if srv == nil {
				return nil, errors.New("server must be set")
			}
			client := apiclient.New(srv.URL, srv.Token, srv.Insecure, srv.Proxy)
			return client, nil
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

func addConfigFlag(root *cobra.Command) {
	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	carapace.Gen(root).FlagCompletion(carapace.ActionMap{
		"config": carapace.ActionFiles("toml"),
	})
}

func addJSONFlag(root *cobra.Command) {
	for _, cmd := range root.Commands() {
		if cmd.HasSubCommands() {
			addJSONFlag(cmd)
		}

		if cmd.Name() == "help" ||
			cmd.Name() == "completion" ||
			strings.HasPrefix(cmd.Name(), "_") {
			continue
		}

		cmd.Flags().BoolVar(&jsonOutput, "json", false, "JSON output")
	}
}

func addContextCommand(root *cobra.Command) {
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

	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"server": carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			if err := parseConfig(os.Args); err != nil {
				return carapace.ActionMessage(err.Error())
			}
			slog.Info(viper.ConfigFileUsed())
			return carapace.ActionValues(maps.Keys(cfg.Servers)...)
		}),
	})

	root.AddCommand(cmd)
}

func parseConfig(args []string) error {
	for i := 1; i < len(args); i++ {
		if args[i-1] == "--config" {
			viper.SetConfigFile(args[i])
			if err := viper.ReadInConfig(); err != nil {
				return err
			}
			if err := viper.Unmarshal(&cfg); err != nil {
				return err
			}
			if err := cfg.ValidateWithContext(context.Background()); err != nil {
				return err
			}
		}
	}
	return nil
}

func initConfig() {
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

	cobra.CheckErr(viper.ReadInConfig())
	cobra.CheckErr(viper.Unmarshal(&cfg))
	cobra.CheckErr(cfg.ValidateWithContext(context.Background()))
}
