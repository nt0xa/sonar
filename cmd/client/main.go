package main

import (
	"bytes"
	"context"
	"log"
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
		cmd.New(nil, nil, nil).Root(&actions.User{}).Execute()
		return
	}

	//
	// URL & token
	//

	var (
		url, token string
		insecure   bool
	)

	if url = os.Getenv("SONAR_API_URL"); url == "" {
		log.Fatal("Empty SONAR_API_URL")
	}

	if token = os.Getenv("SONAR_API_TOKEN"); token == "" {
		log.Fatal("Empty SONAR_API_TOKEN")
	}

	if os.Getenv("SONAR_API_INSECURE") != "" {
		insecure = true
	}

	//
	// API client
	//

	client := apiclient.New(url, token, insecure)

	//
	// User
	//

	user, err := client.UserCurrent(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	//
	// Command
	//

	c := cmd.New(client, resultHandler, nil)

	command := c.Root(user)

	command.Execute()
}

func resultHandler(ctx context.Context, res interface{}) {
	buf := &bytes.Buffer{}

	switch r := res.(type) {
	case actions.PayloadsListResult:
		for _, p := range r {
			payloadTpl.Execute(buf, p)
			break
		}
	}

	color.Print(buf.String())
}
