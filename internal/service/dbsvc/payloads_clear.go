package dbsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/service"
)

// PayloadsClear implements [service.Service].
func (s *Service) PayloadsClear(
	ctx context.Context,
	in service.PayloadsClearInput,
) (service.PayloadsClearOutput, error) {
	c, ok := service.CallerFrom(ctx)
	if !ok {
		return nil, service.Unauthorized()
	}

	payloads, err := s.db.PayloadsDeleteByNamePart(ctx, c.UserID, in.Name)
	if err != nil {
		return nil, err
	}

	out := make([]service.Payload, len(payloads))

	for i, p := range payloads {
		out[i] = *payload(*p)
	}

	return out, nil
}
