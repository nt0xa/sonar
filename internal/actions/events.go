package actions

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/utils/errors"
)

type EventsActions interface {
	EventsList(context.Context, EventsListParams) (EventsListResult, errors.Error)
}

type EventsHandler interface {
	EventsList(context.Context, EventsListResult)
}

type Event struct {
	ID         int64                  `json:"id"`
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
	PayloadName string `err:"payload" path:"payload"`
	Count       uint   `err:"cound" query:"count"`
	After       int64  `err:"after" query:"after"`
	Before      int64  `err:"before" query:"before"`
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

	return cmd, nil
}
