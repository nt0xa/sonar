package remotesvc

import (
	"context"
	"net/http"
	"net/url"

	"github.com/nt0xa/sonar/internal/modules/api2/apimodels"
	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) PayloadsUpdate(
	ctx context.Context,
	in service.PayloadsUpdateInput,
) (*service.PayloadsUpdateOutput, error) {
	req := apimodels.PayloadsUpdateRequest{
		Name:            in.NewName,
		NotifyProtocols: in.NotifyProtocols,
		StoreEvents:     in.StoreEvents,
	}

	var out service.PayloadsUpdateOutput
	if err := s.do(ctx, http.MethodPatch, "/payloads/"+url.PathEscape(in.Name), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
