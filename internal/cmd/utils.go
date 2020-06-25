package cmd

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/utils/errors"
)

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

func RunE(f runEFunc) func(*cobra.Command, []string) error {
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
