package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func (c *command) PayloadsCreate() *cobra.Command {
	var p actions.PayloadsCreateParams

	cmd := &cobra.Command{
		Use:   "new NAME",
		Short: "Create new payload",
		Long:  "Create new payload identified by NAME",
		Args:  OneArg("NAME"),
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			p.Name = args[0]

			res, err := c.actions.PayloadsCreate(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.handler.PayloadsCreate(cmd.Context(), res)

			return nil
		}),
	}

	cmd.Flags().StringSliceVarP(&p.NotifyProtocols, "protocols", "p",
		models.PayloadProtocolsAll, "Protocols to notify")

	return cmd
}

func (c *command) PayloadsList() *cobra.Command {
	var p actions.PayloadsListParams

	cmd := &cobra.Command{
		Use:   "list SUBSTR",
		Short: "List payloads",
		Long:  "List payloads whose NAME contain SUBSTR",
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			if len(args) > 0 {
				p.Name = args[0]
			}

			res, err := c.actions.PayloadsList(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.handler.PayloadsList(cmd.Context(), res)

			return nil
		}),
	}

	return cmd
}

func (c *command) PayloadsUpdate() *cobra.Command {
	var p actions.PayloadsUpdateParams

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

			res, err := c.actions.PayloadsUpdate(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.handler.PayloadsUpdate(cmd.Context(), res)

			return nil
		}),
	}

	cmd.Flags().StringP("name", "n", "", "Payload name")
	cmd.Flags().StringSliceP("protocols", "p", []string{}, "Protocols to notify")

	return cmd
}

func (c *command) PayloadsDelete() *cobra.Command {
	var p actions.PayloadsDeleteParams

	cmd := &cobra.Command{
		Use:   "del NAME",
		Short: "Delete payload",
		Long:  "Delete payload identified by NAME",
		Args:  OneArg("NAME"),
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			p.Name = args[0]

			res, err := c.actions.PayloadsDelete(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.handler.PayloadsDelete(cmd.Context(), res)

			return nil
		}),
	}

	return cmd
}
