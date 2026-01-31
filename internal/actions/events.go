package actions

import (
	"context"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

const (
	EventsListResultID = "events/list"
	EventsGetResultID  = "events/get"
)

type EventsActions interface {
	EventsList(context.Context, EventsListParams) (EventsListResult, errors.Error)
	EventsGet(context.Context, EventsGetParams) (*EventsGetResult, errors.Error)
}

type EventsHandler interface {
	EventsList(context.Context, EventsListResult)
	EventsGet(context.Context, EventsGetResult)
}

type Event struct {
	Index      int64               `json:"index"`
	UUID       string              `json:"uuid"`
	Protocol   string              `json:"protocol"`
	R          string              `json:"r,omitempty"`
	W          string              `json:"w,omitempty"`
	RW         string              `json:"rw,omitempty"`
	Meta       database.EventsMeta `json:"meta"`
	RemoteAddr string              `json:"remoteAddress"`
	ReceivedAt time.Time           `json:"receivedAt"`
}

//
// List
//

type EventsListParams struct {
	PayloadName string `err:"payload" path:"payload" query:"-"`
	Limit       uint   `err:"limit"   query:"limit,omitempty"`
	Offset      uint   `err:"offset"  query:"offset,omitempty"`
}

func (p EventsListParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
	)
}

type EventsListResult []Event

func (r EventsListResult) ResultID() string {
	return EventsListResultID
}

func EventsListCommand(acts *Actions, p *EventsListParams, local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List payload events",
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().UintVarP(&p.Limit, "limit", "l", 10, "Limit")
	cmd.Flags().UintVarP(&p.Offset, "offset", "o", 0, "Offset")

	_ = cmd.RegisterFlagCompletionFunc("payload", completePayloadName(acts))

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

type EventsGetResult struct {
	Event
}

func (r EventsGetResult) ResultID() string {
	return EventsGetResultID
}

func EventsGetCommand(acts *Actions, p *EventsGetParams, local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "get INDEX",
		Short: "Get payload event by INDEX",
		Args:  oneArg("INDEX"),
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")

	_ = cmd.RegisterFlagCompletionFunc("payload", completePayloadName(acts))

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		i, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return errors.Validationf("invalid integer value %q", args[0])
		}
		p.Index = i
		return nil
	}
}
