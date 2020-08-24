package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
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
			p.Name = args[0]

			res, err := c.Actions.CreatePayload(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.ResultHandler(cmd.Context(), res)

			return nil
		}),
	}

	cmd.Flags().StringSliceVarP(&p.NotifyProtocols, "protocols", "p",
		models.PayloadProtocolsAll, "Protocols to notify")

	return cmd
}

func (c *Command) UpdatePayload() *cobra.Command {
	var p actions.UpdatePayloadParams

	cmd := &cobra.Command{
		Use:   "mod NAME",
		Short: "Modify existing payload",
		Long:  "Modify existing payload identified by NAME",
		Args:  OneArg("NAME"),
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			p.Name = args[0]

			if cmd.Flags().Changed("name") {
				newName, _ := cmd.Flags().GetString("name")
				p.NewName = newName
			}

			if cmd.Flags().Changed("protocols") {
				protocols, _ := cmd.Flags().GetStringSlice("protocols")
				p.NotifyProtocols = protocols
			}

			res, err := c.Actions.UpdatePayload(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.ResultHandler(cmd.Context(), res)

			return nil
		}),
	}

	cmd.Flags().StringP("name", "n", "", "Payload name")
	cmd.Flags().StringSliceP("protocols", "p", []string{}, "Protocols to notify")

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
			p.Name = args[0]

			res, err := c.Actions.DeletePayload(cmd.Context(), p)
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
			if len(args) > 0 {
				p.Name = args[0]
			}

			res, err := c.Actions.ListPayloads(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.ResultHandler(cmd.Context(), res)

			return nil
		}),
	}

	return cmd
}
