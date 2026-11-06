package dbsvc

import (
	"cmp"
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// UsersList implements [service.Service].
func (s *Service) UsersList(
	ctx context.Context,
	in service.UsersListInput,
) (service.UsersListOutput, error) {
	c, ok := service.CallerFrom(ctx)
	if !ok {
		return nil, service.Unauthorized()
	}
	if !c.IsAdmin {
		return nil, service.Forbiddenf("admin only")
	}

	perPage := cmp.Or(in.PerPage, 10)
	page := cmp.Or(in.Page, 1)

	users, err := s.db.UsersList(ctx, database.UsersListParams{
		Name:       in.Name,
		PageOffset: int64((page - 1) * perPage),
		PageLimit:  int64(perPage),
	})
	if err != nil {
		return nil, err
	}

	out := make([]service.User, len(users))

	for i, u := range users {
		out[i] = *user(*u)
		// Don't expose other users' API tokens.
		out[i].APIToken = nil
	}

	return out, nil
}
