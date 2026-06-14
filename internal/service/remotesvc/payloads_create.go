package remotesvc

import (
	"context"
	"net/http"

	"github.com/nt0xa/sonar/internal/modules/api/apimodels"
	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) PayloadsCreate(
	ctx context.Context,
	in service.PayloadsCreateInput,
) (*service.PayloadsCreateOutput, error) {
	req := apimodels.PayloadsCreateRequest{
		Name:            in.Name,
		NotifyProtocols: in.NotifyProtocols,
		StoreEvents:     in.StoreEvents,
	}

	var out service.PayloadsCreateOutput
	if err := s.do(ctx, http.MethodPost, "/payloads", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
