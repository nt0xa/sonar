package actionsdb

import (
	"context"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/utils/errors"
)

func (act *dbactions) ProfileGet(ctx context.Context) (*actions.ProfileGetResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.ProfileGetResult{User: User(*u)}, nil
}
