package cmd

import (
	"github.com/spf13/cobra"

	"github.com/russtone/sonar/internal/actions"
)

func (c *Command) DNSRecordsClear(onResult func(actions.Result) error) *cobra.Command {
	var params actions.DNSRecordsClearParams

	cmd, prepareFunc := actions.DNSRecordsClearCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.DNSRecordsClear(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) DNSRecordsCreate(onResult func(actions.Result) error) *cobra.Command {
	var params actions.DNSRecordsCreateParams

	cmd, prepareFunc := actions.DNSRecordsCreateCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.DNSRecordsCreate(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) DNSRecordsDelete(onResult func(actions.Result) error) *cobra.Command {
	var params actions.DNSRecordsDeleteParams

	cmd, prepareFunc := actions.DNSRecordsDeleteCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.DNSRecordsDelete(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) DNSRecordsList(onResult func(actions.Result) error) *cobra.Command {
	var params actions.DNSRecordsListParams

	cmd, prepareFunc := actions.DNSRecordsListCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.DNSRecordsList(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) EventsGet(onResult func(actions.Result) error) *cobra.Command {
	var params actions.EventsGetParams

	cmd, prepareFunc := actions.EventsGetCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.EventsGet(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) EventsList(onResult func(actions.Result) error) *cobra.Command {
	var params actions.EventsListParams

	cmd, prepareFunc := actions.EventsListCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.EventsList(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) HTTPRoutesCreate(onResult func(actions.Result) error) *cobra.Command {
	var params actions.HTTPRoutesCreateParams

	cmd, prepareFunc := actions.HTTPRoutesCreateCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.HTTPRoutesCreate(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) HTTPRoutesDelete(onResult func(actions.Result) error) *cobra.Command {
	var params actions.HTTPRoutesDeleteParams

	cmd, prepareFunc := actions.HTTPRoutesDeleteCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.HTTPRoutesDelete(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) HTTPRoutesList(onResult func(actions.Result) error) *cobra.Command {
	var params actions.HTTPRoutesListParams

	cmd, prepareFunc := actions.HTTPRoutesListCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.HTTPRoutesList(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) PayloadsClear(onResult func(actions.Result) error) *cobra.Command {
	var params actions.PayloadsClearParams

	cmd, prepareFunc := actions.PayloadsClearCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.PayloadsClear(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) PayloadsCreate(onResult func(actions.Result) error) *cobra.Command {
	var params actions.PayloadsCreateParams

	cmd, prepareFunc := actions.PayloadsCreateCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.PayloadsCreate(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) PayloadsDelete(onResult func(actions.Result) error) *cobra.Command {
	var params actions.PayloadsDeleteParams

	cmd, prepareFunc := actions.PayloadsDeleteCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.PayloadsDelete(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) PayloadsList(onResult func(actions.Result) error) *cobra.Command {
	var params actions.PayloadsListParams

	cmd, prepareFunc := actions.PayloadsListCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.PayloadsList(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) PayloadsUpdate(onResult func(actions.Result) error) *cobra.Command {
	var params actions.PayloadsUpdateParams

	cmd, prepareFunc := actions.PayloadsUpdateCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.PayloadsUpdate(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) ProfileGet(onResult func(actions.Result) error) *cobra.Command {
	cmd, prepareFunc := actions.ProfileGetCommand(c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		res, err := c.actions.ProfileGet(cmd.Context())
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) UsersCreate(onResult func(actions.Result) error) *cobra.Command {
	var params actions.UsersCreateParams

	cmd, prepareFunc := actions.UsersCreateCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.UsersCreate(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}

func (c *Command) UsersDelete(onResult func(actions.Result) error) *cobra.Command {
	var params actions.UsersDeleteParams

	cmd, prepareFunc := actions.UsersDeleteCommand(&params, c.options.allowFileAccess)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if prepareFunc != nil {
			if err := prepareFunc(cmd, args); err != nil {
				return err
			}
		}

		if err := params.Validate(); err != nil {
			return err
		}

		res, err := c.actions.UsersDelete(cmd.Context(), params)
		if err != nil {
			return err
		}

		return onResult(res)
	}

	return cmd
}
