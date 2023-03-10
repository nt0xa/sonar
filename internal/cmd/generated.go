package cmd

import (
	"github.com/spf13/cobra"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/utils/errors"
)

func (c *command) DNSRecordsCreate(local bool) *cobra.Command {
	var params actions.DNSRecordsCreateParams

	cmd, prepareFunc := actions.DNSRecordsCreateCommand(&params, local)

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

func (c *command) DNSRecordsDelete(local bool) *cobra.Command {
	var params actions.DNSRecordsDeleteParams

	cmd, prepareFunc := actions.DNSRecordsDeleteCommand(&params, local)

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

func (c *command) DNSRecordsList(local bool) *cobra.Command {
	var params actions.DNSRecordsListParams

	cmd, prepareFunc := actions.DNSRecordsListCommand(&params, local)

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

func (c *command) EventsGet(local bool) *cobra.Command {
	var params actions.EventsGetParams

	cmd, prepareFunc := actions.EventsGetCommand(&params, local)

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

func (c *command) EventsList(local bool) *cobra.Command {
	var params actions.EventsListParams

	cmd, prepareFunc := actions.EventsListCommand(&params, local)

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

func (c *command) HTTPRoutesCreate(local bool) *cobra.Command {
	var params actions.HTTPRoutesCreateParams

	cmd, prepareFunc := actions.HTTPRoutesCreateCommand(&params, local)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.HTTPRoutesCreate(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.HTTPRoutesCreate(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) HTTPRoutesDelete(local bool) *cobra.Command {
	var params actions.HTTPRoutesDeleteParams

	cmd, prepareFunc := actions.HTTPRoutesDeleteCommand(&params, local)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.HTTPRoutesDelete(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.HTTPRoutesDelete(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) HTTPRoutesList(local bool) *cobra.Command {
	var params actions.HTTPRoutesListParams

	cmd, prepareFunc := actions.HTTPRoutesListCommand(&params, local)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}
		if err := params.Validate(); err != nil {
			return errors.Validation(err)
		}

		res, err := c.actions.HTTPRoutesList(cmd.Context(), params)
		if err != nil {
			return err
		}

		c.handler.HTTPRoutesList(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) PayloadsCreate(local bool) *cobra.Command {
	var params actions.PayloadsCreateParams

	cmd, prepareFunc := actions.PayloadsCreateCommand(&params, local)

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

func (c *command) PayloadsDelete(local bool) *cobra.Command {
	var params actions.PayloadsDeleteParams

	cmd, prepareFunc := actions.PayloadsDeleteCommand(&params, local)

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

func (c *command) PayloadsList(local bool) *cobra.Command {
	var params actions.PayloadsListParams

	cmd, prepareFunc := actions.PayloadsListCommand(&params, local)

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

func (c *command) PayloadsUpdate(local bool) *cobra.Command {
	var params actions.PayloadsUpdateParams

	cmd, prepareFunc := actions.PayloadsUpdateCommand(&params, local)

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

func (c *command) ProfileGet(local bool) *cobra.Command {
	cmd, prepareFunc := actions.ProfileGetCommand(local)

	cmd.RunE = RunE(func(cmd *cobra.Command, args []string) errors.Error {

		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		res, err := c.actions.ProfileGet(cmd.Context())
		if err != nil {
			return err
		}

		c.handler.ProfileGet(cmd.Context(), res)

		return nil
	})

	return cmd
}

func (c *command) UsersCreate(local bool) *cobra.Command {
	var params actions.UsersCreateParams

	cmd, prepareFunc := actions.UsersCreateCommand(&params, local)

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

func (c *command) UsersDelete(local bool) *cobra.Command {
	var params actions.UsersDeleteParams

	cmd, prepareFunc := actions.UsersDeleteCommand(&params, local)

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
