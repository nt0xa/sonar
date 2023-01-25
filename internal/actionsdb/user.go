package actionsdb

import (
	"context"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/utils/errors"
)

func (act *dbactions) UserCurrent(ctx context.Context) (actions.UserCurrentResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return User(u), nil
}
