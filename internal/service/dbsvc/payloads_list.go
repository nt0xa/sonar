package dbsvc

import (
	"cmp"
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// PayloadsList implements [service.Service].
func (s *svc) PayloadsList(
	ctx context.Context,
	in service.PayloadsListInput,
) (service.PayloadsListOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, service.ErrUnauthorized
	}

	perPage := cmp.Or(in.PerPage, 10)
	page := cmp.Or(in.Page, 1)

	payloads, err := s.db.PayloadsFindByUserAndName(ctx, database.PayloadsFindByUserAndNameParams{
		UserID: u.ID,
		Name:   in.Name,
		Limit:  int64(perPage),
		Offset: int64((page - 1) * perPage),
	})
	if err != nil {
		return nil, err
	}

	out := make([]service.Payload, len(payloads))

	for i, p := range payloads {
		out[i] = *payload(*p)
	}

	return out, nil
}
