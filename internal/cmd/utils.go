package cmd

import (
	"context"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

type contextKey string

const (
	userKey contextKey = "user"
)

func GetUser(ctx context.Context) (*database.User, error) {
	u, ok := ctx.Value(userKey).(*database.User)
	if !ok {
		return nil, errors.Internalf("no %q key in context", userKey)
	}
	return u, nil
}

func SetUser(ctx context.Context, u *database.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func mapToStruct(src map[string]string, dst interface{}) error {
	c := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           dst,
		WeaklyTypedInput: true,
	}
	decoder, err := mapstructure.NewDecoder(c)
	if err != nil {
		return err
	}

	return decoder.Decode(src)
}

type runEFunc func(*cobra.Command, []string) errors.Error

// runE is wrapper for better type checking
func runE(f runEFunc) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return f(cmd, args)
	}
}

func OneArg(name string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.Validationf("argument %q is required", name)
		}
		return nil
	}
}
