package remotesvc

import (
	"context"
	"net/http"
	"net/url"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) DNSRecordsClear(
	ctx context.Context,
	in service.DNSRecordsClearInput,
) (service.DNSRecordsClearOutput, error) {
	q := url.Values{}
	if in.Name != "" {
		q.Set("name", in.Name)
	}

	path := withQuery("/dns-records/"+url.PathEscape(in.PayloadName), q)

	var out service.DNSRecordsClearOutput
	if err := s.do(ctx, http.MethodDelete, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
