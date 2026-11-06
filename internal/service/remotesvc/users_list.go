package remotesvc

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) UsersList(
	ctx context.Context,
	in service.UsersListInput,
) (service.UsersListOutput, error) {
	q := url.Values{}
	if in.Name != "" {
		q.Set("name", in.Name)
	}
	if in.Page != 0 {
		q.Set("page", strconv.FormatUint(uint64(in.Page), 10))
	}
	if in.PerPage != 0 {
		q.Set("perPage", strconv.FormatUint(uint64(in.PerPage), 10))
	}

	var out service.UsersListOutput
	if err := s.do(ctx, http.MethodGet, withQuery("/users", q), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
