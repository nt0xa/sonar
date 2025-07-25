package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nt0xa/sonar/internal/cmd/server"
)

func main() {
	ctx := context.Background()

	err := server.Run(
		ctx,
		os.Stdout,
		os.Stderr,
		server.ConfigDefaults,
		[]byte{},
		os.Environ,
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run server: %s\n", err)
		os.Exit(1)
	}
}
