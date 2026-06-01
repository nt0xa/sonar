package service

import (
	"context"

	"github.com/nt0xa/sonar/internal/types"
)

// PayloadsClear implements [types.Service].
func (s *service) PayloadsClear(
	ctx context.Context,
	in types.PayloadsClearInput,
) (types.PayloadsClearOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, types.ErrUnauthorized
	}

	payloads, err := s.db.PayloadsDeleteByNamePart(ctx, u.ID, in.Name)
	if err != nil {
		return nil, err
	}

	out := make([]types.Payload, len(payloads))

	for i, p := range payloads {
		out[i] = *payload(*p)
	}

	return out, nil
}
