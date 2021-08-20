package main

import (
	"context"
	"net/url"
	"os"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/cmd"
	"github.com/bi-zone/sonar/internal/modules/api/apiclient"
	"github.com/bi-zone/sonar/internal/utils/slice"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
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
		root.AddCommand(completionCmd)
		root.Execute()
		return
	}

	//
	// URL & token
	//

	var (
		baseURL, token string
		insecure       bool
		proxy          *string
	)

	if baseURL = os.Getenv("SONAR_URL"); baseURL == "" {
		fatal("Empty SONAR_URL")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		fatal(err)
	}

	if token = os.Getenv("SONAR_TOKEN"); token == "" {
		fatal("Empty SONAR_TOKEN")
	}

	if os.Getenv("SONAR_INSECURE") != "" {
		insecure = true
	}

	if p := os.Getenv("SONAR_PROXY"); p != "" {
		proxy = &p
	}

	//
	// API client
	//

	client := apiclient.New(baseURL, token, insecure, proxy)

	//
	// User
	//

	user, err := client.UserCurrent(context.Background())
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
		handler = &terminalHandler{u.Hostname()}
	}

	root := cmd.New(client, handler, nil).Root(user, true)
	root.AddCommand(completionCmd)
	addJSONFlag(root)
	root.SilenceErrors = true
	root.SilenceUsage = true

	if err := root.Execute(); err != nil {
		fatal(err)
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

		cmd.Flags().Bool("json", false, "JSON output")
	}

}

func fatal(data interface{}) {
	color.Danger.Println(data)
	os.Exit(1)
}
