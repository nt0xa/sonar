package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func (c *Command) CreatePayload() *cobra.Command {
	var p actions.CreatePayloadParams

	cmd := &cobra.Command{
		Use:   "new NAME",
		Short: "Create new payload",
		Long:  "Create new payload identified by NAME",
		Args:  OneArg("NAME"),
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			u, err := GetUser(cmd.Context())
			if err != nil {
				return err
			}

			p.Name = args[0]

			res, err := c.Actions.CreatePayload(u, p)
			if err != nil {
				return err
			}

			c.ResultHandler(cmd.Context(), res)

			return nil
		}),
	}

	return cmd
}

func (c *Command) DeletePayload() *cobra.Command {
	var p actions.DeletePayloadParams

	cmd := &cobra.Command{
		Use:   "del NAME",
		Short: "Delete payload",
		Long:  "Delete payload identified by NAME",
		Args:  OneArg("NAME"),
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			u, err := GetUser(cmd.Context())
			if err != nil {
				return err
			}

			p.Name = args[0]

			res, err := c.Actions.DeletePayload(u, p)
			if err != nil {
				return err
			}

			c.ResultHandler(cmd.Context(), res)

			return nil
		}),
	}

	return cmd
}

func (c *Command) ListPayloads() *cobra.Command {
	var p actions.ListPayloadsParams

	cmd := &cobra.Command{
		Use:   "list SUBSTR",
		Short: "List payloads",
		Long:  "List payloads whose NAME contain SUBSTR",
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			u, err := GetUser(cmd.Context())
			if err != nil {
				return err
			}

			if len(args) > 0 {
				p.Name = args[0]
			}

			res, err := c.Actions.ListPayloads(u, p)
			if err != nil {
				return err
			}

			c.ResultHandler(cmd.Context(), res)

			return nil
		}),
	}

	return cmd
}
