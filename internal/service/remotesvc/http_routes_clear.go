package remotesvc

import (
	"context"
	"net/http"
	"net/url"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) HTTPRoutesClear(
	ctx context.Context,
	in service.HTTPRoutesClearInput,
) (service.HTTPRoutesClearOutput, error) {
	q := url.Values{}
	if in.Path != "" {
		q.Set("path", in.Path)
	}

	path := withQuery("/http-routes/"+url.PathEscape(in.PayloadName), q)

	var out service.HTTPRoutesClearOutput
	if err := s.do(ctx, http.MethodDelete, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
