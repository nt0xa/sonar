package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/adrg/xdg"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"

	"github.com/nt0xa/sonar/internal/cmd"
	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/internal/service/remotesvc"
	"github.com/nt0xa/sonar/internal/templates"
)

// lazyService lets the command tree be built before the config (and thus the
// remote service URL/token) is known: the embedded service.Service is nil at
// construction and set in PersistentPreRunE once config is loaded. Method
// promotion reads the embedded value at call time.
type lazyService struct {
	service.Service
}

func main() {
	var (
		cfg        Config
		cfgFile    string
		jsonOutput bool
	)

	svc := &lazyService{}

	c := cmd.New(
		svc,
		cmd.AllowFileAccess(true),
		cmd.PreExec(func(root *cobra.Command) {
			root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
				// Skip config & service initialization for "help" and "completion" commands.
				if isHelpOrCompletion(cmd.CommandPath()) {
					return nil
				}

				// We have to find the flag value ourselves because flags are not parsed on completion.
				if err := initConfig(findFlagValue("config", os.Args), &cfg); err != nil {
					return err
				}

				s, err := newService(cfg)
				if err != nil {
					return err
				}
				svc.Service = s

				return nil
			}

			// Flags, commands...
			root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
			jsonFlag(root, &jsonOutput)
			contextCmd(root, &cfg)
		}),
	)

	res, err := c.Exec(context.Background(), os.Args[1:])
	cobra.CheckErr(err)

	switch v := res.(type) {
	case string:
		// Help/usage/completion text produced by cobra (no leaf result).
		_, _ = fmt.Fprint(os.Stdout, v)
	default:
		if jsonOutput {
			cobra.CheckErr(json.NewEncoder(os.Stdout).Encode(v))
			return
		}
		tmpl := templates.New(cfg.Server().BaseURL().Hostname(),
			templates.Default(
				templates.HTMLEscape(false),
				templates.Markup(templates.Bold("<bold>", "</>"))),
		)
		s, err := tmpl.RenderResult(v)
		cobra.CheckErr(err)
		color.Fprint(os.Stdout, s)
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
	_ = viper.BindPFlag("context.server", cmd.Flags().Lookup("server"))

	_ = cmd.RegisterFlagCompletionFunc("server", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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

	if p := cfg.Validate(); !p.Ok() {
		return fmt.Errorf("config validation failed: %w", p)
	}

	return nil
}

func newService(cfg Config) (service.Service, error) {
	srv := cfg.Server()
	if srv == nil {
		return nil, errors.New("server must be set")
	}

	transport := &http.Transport{}
	if srv.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if u := srv.ProxyURL(); u != nil {
		transport.Proxy = http.ProxyURL(u)
	}

	return remotesvc.New(srv.URL, srv.Token, &http.Client{Transport: transport}), nil
}

func findFlagValue(f string, args []string) string {
	for i := 1; i < len(args); i++ {
		if args[i-1] == "--"+f {
			return args[i]
		}
	}
	return ""
}

func isHelpOrCompletion(path string) bool {
	parts := strings.Split(path, " ")
	return slices.Contains(parts, "completion") || slices.Contains(parts, "help")
}
