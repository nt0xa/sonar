package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

func Run(
	ctx context.Context,
	environFunc func() []string,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg, err := GetConfig(nil, nil, environFunc)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	fmt.Printf("cfg = %+v\n", cfg)

	return nil
}
