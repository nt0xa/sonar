package main

import (
	"context"
	"os"
	"reflect"

	"github.com/adrg/xdg"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/cmd"
	"github.com/russtone/sonar/internal/modules/api/apiclient"
	"github.com/russtone/sonar/internal/utils/slice"
)

func main() {

	// Allow "help" and "completion" commands to execute without any
	// API requests.
	if len(os.Args) > 1 &&
		(os.Args[1] == "help" ||
			os.Args[1] == "completion" ||
			slice.StringsContains(os.Args, "-h") ||
			slice.StringsContains(os.Args, "--help")) {
		root := cmd.New(nil, nil, nil).Root(&actions.User{}, true)
		addJSONFlag(root)
		root.Execute()
		return
	}

	//
	// Config
	//

	var cfg Config

	configFilePath, err := xdg.ConfigFile("sonar/config.toml")
	if err != nil {
		fatalf("Fail to obtain config file path: %v", err)
	}
	viper.SetConfigFile(configFilePath)

	if err := viper.ReadInConfig(); err != nil {
		fatalf("Fail to read config: %v", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		fatalf("Fail to unmarshal config: %v", err)
	}

	if err := cfg.ValidateWithContext(context.Background()); err != nil {
		fatalf("Config validation failed: %v", err)
	}

	srv, ok := cfg.Servers[cfg.Context.Server]
	if !ok {
		fatalf("Invalid server %q", cfg.Context.Server)
	}

	//
	// API client
	//

	client := apiclient.New(srv.URL, srv.Token, srv.Insecure, srv.Proxy)

	//
	// User
	//

	user, err := client.ProfileGet(context.Background())
	if err != nil {
		fatal(err)
	}

	//
	// Command
	//

	var handler actions.ResultHandler

	// Args are is not yet parsed so just seach for "--json".
	if slice.StringsContains(os.Args, "--json") {
		handler = &jsonHandler{os.Stdout}
	} else {
		handler = &terminalHandler{srv.BaseURL().Hostname()}
	}

	root := cmd.New(client, handler, nil).Root(user, true)
	addJSONFlag(root)
	addContextCmd(&cfg, root)
	root.SilenceErrors = true
	root.SilenceUsage = true

	if err := root.Execute(); err != nil {
		fatal(err)
	}
}

var jsonOutput bool

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

func addContextCmd(cfg *Config, root *cobra.Command) {
	var server string

	cmd := &cobra.Command{
		Use:   "ctx",
		Short: "Change current context parameters",
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := viper.Unmarshal(&cfg); err != nil {
				fatalf("Fail to unmarshal config: %v", err)
			}

			if err := cfg.ValidateWithContext(context.Background()); err != nil {
				fatalf("Config validation failed: %v", err)
			}

			if err := viper.WriteConfig(); err != nil {
				fatalf("Fail to update config: %v", err)
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

func fatal(data interface{}) {
	color.Danger.Println(data)
	os.Exit(1)
}

func fatalf(format string, a ...interface{}) {
	color.Danger.Printf(format+"\n", a...)
	os.Exit(1)
}
