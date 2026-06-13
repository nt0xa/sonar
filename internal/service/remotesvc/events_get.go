package remotesvc

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) EventsGet(
	ctx context.Context,
	in service.EventsGetInput,
) (*service.EventsGetOutput, error) {
	path := "/events/" + url.PathEscape(in.PayloadName) + "/" + strconv.FormatInt(in.Index, 10)

	var out service.EventsGetOutput
	if err := s.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
