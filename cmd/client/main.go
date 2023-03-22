package main

import (
	"context"
	"encoding/json"
	"html/template"
	"os"
	"reflect"
	"strings"

	"github.com/adrg/xdg"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/cmd"
	"github.com/russtone/sonar/internal/modules/api/apiclient"
	"github.com/russtone/sonar/internal/results"
	"github.com/russtone/sonar/internal/utils/slice"
)

func init() {
	validation.ErrorTag = "err"
}

func main() {

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
	// Command
	//

	var handler actions.ResultHandler

	// Args are is not yet parsed so just seach for "--json".
	if slice.StringsContains(os.Args, "--json") {
		handler = &results.JSON{Encoder: json.NewEncoder(os.Stdout)}
	} else {
		handler = &results.Text{
			Templates: results.DefaultTemplates(results.TemplateOptions{
				Markup: map[string]string{
					"<bold>":   "<bold>",
					"</bold>":  "</>",
					"<code>":   "",
					"</code>":  "",
					"<pre>":    "",
					"</pre>":   "",
					"<error>":  "<fg=red;op=bold>",
					"</error>": "</>",
				},
				ExtraFuncs: template.FuncMap{
					"domain": func() string {
						return srv.BaseURL().Hostname()
					},
				},
			}),
			OnText: func(ctx context.Context, id, message string) {
				if !strings.HasSuffix(message, "\n") {
					message += "\n"
				}
				color.Print(message)
			},
		}
	}

	c := cmd.New(
		client,
		handler,
		cmd.Local(),
		cmd.PreExec(preExec(&cfg)),
	)

	c.Exec(context.Background(), os.Args[1:])
}

var jsonOutput bool

func preExec(cfg *Config) func(context.Context, *cobra.Command) {
	return func(ctx context.Context, root *cobra.Command) {
		addJSONFlag(root)
		addContextCmd(cfg, root)
		root.SilenceErrors = true
		root.SilenceUsage = true
	}
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

func addContextCmd(cfg *Config, root *cobra.Command) {
	var server string

	cmd := &cobra.Command{
		Use:   "ctx",
		Short: "Change current context parameters",
		Run: func(cmd *cobra.Command, args []string) {

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
