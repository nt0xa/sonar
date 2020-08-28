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
)

func main() {

	// Allow "help" and "completion" commands to execute without any
	// API requests.
	if len(os.Args) > 1 &&
		(os.Args[1] == "help" ||
			os.Args[1] == "completion" ||
			slice.StringsContains(os.Args, "-h") ||
			slice.StringsContains(os.Args, "--help")) {
		root := cmd.New(nil, nil, nil).Root(&actions.User{})
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
	)

	if baseURL = os.Getenv("SONAR_API_URL"); baseURL == "" {
		fatal("Empty SONAR_API_URL")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		fatal(err)
	}

	if token = os.Getenv("SONAR_API_TOKEN"); token == "" {
		fatal("Empty SONAR_API_TOKEN")
	}

	if os.Getenv("SONAR_API_INSECURE") != "" {
		insecure = true
	}

	//
	// API client
	//

	client := apiclient.New(baseURL, token, insecure)

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

	root := cmd.New(client, &handler{u.Hostname()}, nil).Root(user)
	root.AddCommand(completionCmd)
	root.SilenceErrors = true
	root.SilenceUsage = true
	if err := root.Execute(); err != nil {
		fatal(err)
	}
}

func fatal(data interface{}) {
	color.Danger.Println(data)
	os.Exit(1)
}
