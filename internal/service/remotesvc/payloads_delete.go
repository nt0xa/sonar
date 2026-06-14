package remotesvc

import (
	"context"
	"net/http"
	"net/url"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) PayloadsDelete(
	ctx context.Context,
	in service.PayloadsDeleteInput,
) (*service.PayloadsDeleteOutput, error) {
	var out service.PayloadsDeleteOutput
	if err := s.do(ctx, http.MethodDelete, "/payloads/"+url.PathEscape(in.Name), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
