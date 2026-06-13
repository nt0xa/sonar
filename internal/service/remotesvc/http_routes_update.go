package remotesvc

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/nt0xa/sonar/internal/modules/api2/apimodels"
	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) HTTPRoutesUpdate(
	ctx context.Context,
	in service.HTTPRoutesUpdateInput,
) (*service.HTTPRoutesUpdateOutput, error) {
	req := apimodels.HTTPRoutesUpdateRequest{
		Method:    in.Method,
		Path:      in.Path,
		Code:      in.Code,
		Headers:   in.Headers,
		Body:      in.Body,
		IsDynamic: in.IsDynamic,
	}

	path := "/http-routes/" + url.PathEscape(in.Payload) + "/" + strconv.FormatInt(in.Index, 10)

	var out service.HTTPRoutesUpdateOutput
	if err := s.do(ctx, http.MethodPatch, path, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
