package remotesvc

import (
	"context"
	"net/http"
	"net/url"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) HTTPRoutesList(
	ctx context.Context,
	in service.HTTPRoutesListInput,
) (service.HTTPRoutesListOutput, error) {
	var out service.HTTPRoutesListOutput
	if err := s.do(ctx, http.MethodGet, "/http-routes/"+url.PathEscape(in.PayloadName), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
