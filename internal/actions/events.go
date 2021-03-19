package actions

import (
	"context"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/utils/errors"
)

type EventsActions interface {
	EventsList(context.Context, EventsListParams) (EventsListResult, errors.Error)
	EventsGet(context.Context, EventsGetParams) (EventsGetResult, errors.Error)
}

type EventsHandler interface {
	EventsList(context.Context, EventsListResult)
	EventsGet(context.Context, EventsGetResult)
}

type Event struct {
	Index      int64                  `json:"index"`
	Protocol   string                 `json:"protocol"`
	R          string                 `json:"r,omitempty"`
	W          string                 `json:"w,omitempty"`
	RW         string                 `json:"rw,omitempty"`
	Meta       map[string]interface{} `json:"meta"`
	RemoteAddr string                 `json:"remoteAddress"`
	ReceivedAt time.Time              `json:"receivedAt"`
}

//
// List
//

type EventsListParams struct {
	PayloadName string `err:"payload" path:"payload" query:"-"`
	Count       uint   `err:"cound"   query:"count,omitempty"`
	After       int64  `err:"after"   query:"after,omitempty"`
	Before      int64  `err:"before"  query:"before,omitempty"`
	Reverse     bool   `err:"reverse" query:"reverse,omitempty"`
}

func (p EventsListParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
	)
}

type EventsListResult []*Event

func EventsListCommand(p *EventsListParams) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List payload events",
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().UintVarP(&p.Count, "count", "c", 10, "Count of events")
	cmd.Flags().Int64VarP(&p.After, "after", "a", 0, "After ID")
	cmd.Flags().Int64VarP(&p.Before, "before", "b", 0, "Before ID")
	cmd.Flags().BoolVarP(&p.Reverse, "reverse", "r", false, "List events in reversed order")

	return cmd, nil
}

//
// GetOne
//

type EventsGetParams struct {
	PayloadName string `err:"payload" path:"payload"`
	Index       int64  `err:"index"   path:"index"`
}

func (p EventsGetParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
		validation.Field(&p.Index, validation.Required),
	)
}

type EventsGetResult *Event

func EventsGetCommand(p *EventsGetParams) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "get INDEX",
		Short: "Get payload event by INDEX",
		Args:  oneArg("INDEX"),
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		i, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return errors.Validationf("invalid integer value %q", args[0])
		}
		p.Index = i
		return nil
	}
}
