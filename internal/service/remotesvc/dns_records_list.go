package remotesvc

import (
	"context"
	"net/http"
	"net/url"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) DNSRecordsList(
	ctx context.Context,
	in service.DNSRecordsListInput,
) (service.DNSRecordsListOutput, error) {
	var out service.DNSRecordsListOutput
	if err := s.do(ctx, http.MethodGet, "/dns-records/"+url.PathEscape(in.PayloadName), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
