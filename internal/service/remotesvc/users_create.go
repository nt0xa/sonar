package remotesvc

import (
	"context"
	"net/http"

	"github.com/nt0xa/sonar/internal/modules/api/apimodels"
	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) UsersCreate(
	ctx context.Context,
	in service.UsersCreateInput,
) (*service.UsersCreateOutput, error) {
	req := apimodels.UsersCreateRequest{
		Name:       in.Name,
		APIToken:   in.APIToken,
		TelegramID: in.TelegramID,
		LarkID:     in.LarkID,
		SlackID:    in.SlackID,
		IsAdmin:    in.IsAdmin,
	}

	var out service.UsersCreateOutput
	if err := s.do(ctx, http.MethodPost, "/users", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
