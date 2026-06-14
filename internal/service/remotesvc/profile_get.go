package remotesvc

import (
	"context"
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) ProfileGet(ctx context.Context) (*service.ProfileGetOutput, error) {
	var out service.ProfileGetOutput
	if err := s.do(ctx, http.MethodGet, "/profile", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
