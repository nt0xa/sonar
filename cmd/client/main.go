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
	c := cmd.New(
		nil,
		cmd.AllowFileAccess(true),
		cmd.PreExec(func(root *cobra.Command) {
			addConfigFlag(root)
			addJSONFlag(root)
			addContextCommand(root)
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
}

func addJSONFlag(root *cobra.Command) {
	for _, cmd := range root.Commands() {
		if cmd.HasSubCommands() {
			addJSONFlag(cmd)
		}

		if cmd.Name() == "help" || cmd.Name() == "completion" {
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

	root.AddCommand(cmd)
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
