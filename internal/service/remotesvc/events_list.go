package remotesvc

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) EventsList(
	ctx context.Context,
	in service.EventsListInput,
) (service.EventsListOutput, error) {
	q := url.Values{}
	if in.Limit != 0 {
		q.Set("limit", strconv.FormatUint(uint64(in.Limit), 10))
	}
	if in.Offset != 0 {
		q.Set("offset", strconv.FormatUint(uint64(in.Offset), 10))
	}

	path := withQuery("/events/"+url.PathEscape(in.PayloadName), q)

	var out service.EventsListOutput
	if err := s.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
