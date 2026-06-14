package remotesvc

import (
	"context"
	"net/http"
	"net/url"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) PayloadsClear(
	ctx context.Context,
	in service.PayloadsClearInput,
) (service.PayloadsClearOutput, error) {
	q := url.Values{}
	if in.Name != "" {
		q.Set("name", in.Name)
	}

	var out service.PayloadsClearOutput
	if err := s.do(ctx, http.MethodDelete, withQuery("/payloads", q), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
