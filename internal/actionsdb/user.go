package actionsdb

import (
	"context"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func (act *dbactions) ProfileGet(ctx context.Context) (*actions.ProfileGetResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if u == nil {
		return nil, errors.Unauthorized()
	}

	return &actions.ProfileGetResult{User: User(*u)}, nil
}
