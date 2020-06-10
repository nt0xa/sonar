package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func CreatePayloadCmd(acts actions.Actions, handler ResultHandler) *cobra.Command {
	var p actions.CreatePayloadParams

	cmd := &cobra.Command{
		Use:   "new NAME",
		Short: "Create new payload",
		Long:  "Create new payload identified by NAME",
		Args:  OneArg("NAME"),
		PreRunE: runE(func(cmd *cobra.Command, args []string) errors.Error {
			p.Name = args[0]
			if err := p.Validate(); err != nil {
				return errors.Validation(err)
			}
			return nil
		}),
		RunE: runE(func(cmd *cobra.Command, args []string) errors.Error {
			u, err := GetUser(cmd.Context())
			if err != nil {
				return err
			}

			res, err := acts.CreatePayload(u, p)
			if err != nil {
				return err
			}

			handler(cmd.Context(), res)

			return nil
		}),
	}

	return cmd
}

func DeletePayloadCmd(acts actions.Actions, handler ResultHandler) *cobra.Command {
	var p actions.DeletePayloadParams

	cmd := &cobra.Command{
		Use:   "del NAME",
		Short: "Delete payload",
		Long:  "Delete payload identified by NAME",
		Args:  OneArg("NAME"),
		PreRunE: runE(func(cmd *cobra.Command, args []string) errors.Error {
			p.Name = args[0]
			if err := p.Validate(); err != nil {
				return errors.Validation(err)
			}
			return nil
		}),
		RunE: runE(func(cmd *cobra.Command, args []string) errors.Error {
			u, err := GetUser(cmd.Context())
			if err != nil {
				return err
			}

			res, err := acts.DeletePayload(u, p)
			if err != nil {
				return err
			}

			handler(cmd.Context(), res)

			return nil
		}),
	}

	return cmd
}

func ListPayloadCmd(acts actions.Actions, handler ResultHandler) *cobra.Command {
	var p actions.ListPayloadsParams

	cmd := &cobra.Command{
		Use:   "list SUBSTR",
		Short: "List payloads",
		Long:  "List payloads whose NAME contain SUBSTR",
		PreRunE: runE(func(cmd *cobra.Command, args []string) errors.Error {
			if len(args) > 0 {
				p.Name = args[0]
			}

			if err := p.Validate(); err != nil {
				return errors.Validation(err)
			}
			return nil
		}),
		RunE: runE(func(cmd *cobra.Command, args []string) errors.Error {
			u, err := GetUser(cmd.Context())
			if err != nil {
				return err

			}

			res, err := acts.ListPayloads(u, p)
			if err != nil {
				return err
			}

			handler(cmd.Context(), res)

			return nil
		}),
	}

	return cmd
}
