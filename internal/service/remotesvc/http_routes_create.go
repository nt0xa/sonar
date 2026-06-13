package remotesvc

import (
	"context"
	"net/http"

	"github.com/nt0xa/sonar/internal/modules/api2/apimodels"
	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) HTTPRoutesCreate(
	ctx context.Context,
	in service.HTTPRoutesCreateInput,
) (*service.HTTPRoutesCreateOutput, error) {
	req := apimodels.HTTPRoutesCreateRequest{
		PayloadName: in.PayloadName,
		Method:      in.Method,
		Path:        in.Path,
		Code:        in.Code,
		Headers:     in.Headers,
		Body:        in.Body,
		IsDynamic:   in.IsDynamic,
	}

	var out service.HTTPRoutesCreateOutput
	if err := s.do(ctx, http.MethodPost, "/http-routes", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
