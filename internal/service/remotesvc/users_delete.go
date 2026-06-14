package remotesvc

import (
	"context"
	"net/http"
	"net/url"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) UsersDelete(
	ctx context.Context,
	in service.UsersDeleteInput,
) (*service.UsersDeleteOutput, error) {
	var out service.UsersDeleteOutput
	if err := s.do(ctx, http.MethodDelete, "/users/"+url.PathEscape(in.Name), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
