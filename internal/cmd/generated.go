package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func (c *command) DNSRecordsCreate() *cobra.Command {
	var params actions.DNSRecordsCreateParams

	cmd, prepareFunc := actions.DNSRecordsCreateCommand(&params)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.DNSRecordsCreate(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.DNSRecordsCreate(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) DNSRecordsDelete() *cobra.Command {
	var params actions.DNSRecordsDeleteParams

	cmd, prepareFunc := actions.DNSRecordsDeleteCommand(&params)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.DNSRecordsDelete(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.DNSRecordsDelete(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) DNSRecordsList() *cobra.Command {
	var params actions.DNSRecordsListParams

	cmd, prepareFunc := actions.DNSRecordsListCommand(&params)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.DNSRecordsList(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.DNSRecordsList(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) EventsGet() *cobra.Command {
	var params actions.EventsGetParams

	cmd, prepareFunc := actions.EventsGetCommand(&params)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.EventsGet(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.EventsGet(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) EventsList() *cobra.Command {
	var params actions.EventsListParams

	cmd, prepareFunc := actions.EventsListCommand(&params)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.EventsList(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.EventsList(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) PayloadsCreate() *cobra.Command {
	var params actions.PayloadsCreateParams

	cmd, prepareFunc := actions.PayloadsCreateCommand(&params)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.PayloadsCreate(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.PayloadsCreate(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) PayloadsDelete() *cobra.Command {
	var params actions.PayloadsDeleteParams

	cmd, prepareFunc := actions.PayloadsDeleteCommand(&params)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.PayloadsDelete(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.PayloadsDelete(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) PayloadsList() *cobra.Command {
	var params actions.PayloadsListParams

	cmd, prepareFunc := actions.PayloadsListCommand(&params)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.PayloadsList(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.PayloadsList(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) PayloadsUpdate() *cobra.Command {
	var params actions.PayloadsUpdateParams

	cmd, prepareFunc := actions.PayloadsUpdateCommand(&params)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.PayloadsUpdate(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.PayloadsUpdate(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) UserCurrent() *cobra.Command {
	cmd, prepareFunc := actions.UserCurrentCommand()

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		res, err := c.actions.UserCurrent(cmd.Context())
		if err != nil {
			return err
		}

		c.handler.UserCurrent(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) UsersCreate() *cobra.Command {
	var params actions.UsersCreateParams

	cmd, prepareFunc := actions.UsersCreateCommand(&params)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.UsersCreate(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.UsersCreate(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) UsersDelete() *cobra.Command {
	var params actions.UsersDeleteParams

	cmd, prepareFunc := actions.UsersDeleteCommand(&params)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.UsersDelete(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.UsersDelete(cmd.Context(), res)

		return nil
	})

	return cmd
}
