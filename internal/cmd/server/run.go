package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
)

func Run(
	ctx context.Context,
	stdout io.Writer,
	configDefaults map[string]any,
	configContents []byte,
	environFunc func() []string,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg, err := GetConfig(
		configDefaults,
		configContents,
		environFunc,
	)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	fmt.Printf("cfg = %+v\n", cfg)

	return nil
}
