package actions

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"

	"github.com/russtone/sonar/internal/utils/errors"
)

func oneArg(name string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.Validationf("argument %q is required", name)
		}
		return nil
	}
}

func atLeastOneArg(name string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.Validationf("arguments %q is required", name)
		}
		return nil
	}
}

func quoteAndJoin(values []string) string {
	s := ""

	for i, v := range values {
		s += fmt.Sprintf("%q", v)
		if i != len(values)-1 {
			s += ", "
		}
	}

	return s
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
