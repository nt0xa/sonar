package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/utils/errors"
)

type runEFunc func(*cobra.Command, []string) errors.Error

func RunE(f runEFunc) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return f(cmd, args)
	}
}
