package dbsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/service"
)

// PayloadsClear implements [service.Service].
func (s *svc) PayloadsClear(
	ctx context.Context,
	in service.PayloadsClearInput,
) (service.PayloadsClearOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, service.ErrUnauthorized
	}

	payloads, err := s.db.PayloadsDeleteByNamePart(ctx, u.ID, in.Name)
	if err != nil {
		return nil, err
	}

	out := make([]service.Payload, len(payloads))

	for i, p := range payloads {
		out[i] = *payload(*p)
	}

	return out, nil
}
