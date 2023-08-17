package cmd

import (
	"github.com/spf13/cobra"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/utils/errors"
)

func (c *Command) DNSRecordsCreate() *cobra.Command {
	var params actions.DNSRecordsCreateParams

	cmd, prepareFunc := actions.DNSRecordsCreateCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.DNSRecordsCreate(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) DNSRecordsDelete() *cobra.Command {
	var params actions.DNSRecordsDeleteParams

	cmd, prepareFunc := actions.DNSRecordsDeleteCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.DNSRecordsDelete(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) DNSRecordsList() *cobra.Command {
	var params actions.DNSRecordsListParams

	cmd, prepareFunc := actions.DNSRecordsListCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.DNSRecordsList(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) EventsGet() *cobra.Command {
	var params actions.EventsGetParams

	cmd, prepareFunc := actions.EventsGetCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.EventsGet(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) EventsList() *cobra.Command {
	var params actions.EventsListParams

	cmd, prepareFunc := actions.EventsListCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.EventsList(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) HTTPRoutesCreate() *cobra.Command {
	var params actions.HTTPRoutesCreateParams

	cmd, prepareFunc := actions.HTTPRoutesCreateCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.HTTPRoutesCreate(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) HTTPRoutesDelete() *cobra.Command {
	var params actions.HTTPRoutesDeleteParams

	cmd, prepareFunc := actions.HTTPRoutesDeleteCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.HTTPRoutesDelete(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) HTTPRoutesList() *cobra.Command {
	var params actions.HTTPRoutesListParams

	cmd, prepareFunc := actions.HTTPRoutesListCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.HTTPRoutesList(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) PayloadsCreate() *cobra.Command {
	var params actions.PayloadsCreateParams

	cmd, prepareFunc := actions.PayloadsCreateCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.PayloadsCreate(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) PayloadsDelete() *cobra.Command {
	var params actions.PayloadsDeleteParams

	cmd, prepareFunc := actions.PayloadsDeleteCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.PayloadsDelete(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) PayloadsList() *cobra.Command {
	var params actions.PayloadsListParams

	cmd, prepareFunc := actions.PayloadsListCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.PayloadsList(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) PayloadsUpdate() *cobra.Command {
	var params actions.PayloadsUpdateParams

	cmd, prepareFunc := actions.PayloadsUpdateCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.PayloadsUpdate(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) ProfileGet() *cobra.Command {
	cmd, prepareFunc := actions.ProfileGetCommand(c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}

		res, err := c.actions.ProfileGet(cmd.Context())
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) UsersCreate() *cobra.Command {
	var params actions.UsersCreateParams

	cmd, prepareFunc := actions.UsersCreateCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.UsersCreate(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}

func (c *Command) UsersDelete() *cobra.Command {
	var params actions.UsersDeleteParams

	cmd, prepareFunc := actions.UsersDeleteCommand(&params, c.local)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				c.handler.OnResult(cmd.Context(), actions.Error(err))
				return
			}
		}
		if err := params.Validate(); err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(errors.Validation(err)))
			return
		}

		res, err := c.actions.UsersDelete(cmd.Context(), params)
		if err != nil {
			c.handler.OnResult(cmd.Context(), actions.Error(err))
			return
		}

		c.handler.OnResult(cmd.Context(), res)

		return
	}

	return cmd
}
