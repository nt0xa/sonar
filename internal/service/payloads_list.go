package service

import (
	"cmp"
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

// PayloadsList implements [types.Service].
func (s *service) PayloadsList(
	ctx context.Context,
	in types.PayloadsListInput,
) (types.PayloadsListOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, types.ErrUnauthorized
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

	out := make([]types.Payload, len(payloads))

	for i, p := range payloads {
		out[i] = *payload(*p)
	}

	return out, nil
}
