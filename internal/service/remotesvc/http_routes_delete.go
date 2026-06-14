package remotesvc

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) HTTPRoutesDelete(
	ctx context.Context,
	in service.HTTPRoutesDeleteInput,
) (*service.HTTPRoutesDeleteOutput, error) {
	path := "/http-routes/" + url.PathEscape(in.PayloadName) + "/" + strconv.FormatInt(in.Index, 10)

	var out service.HTTPRoutesDeleteOutput
	if err := s.do(ctx, http.MethodDelete, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
