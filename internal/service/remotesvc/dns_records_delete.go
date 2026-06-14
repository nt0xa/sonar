package remotesvc

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) DNSRecordsDelete(
	ctx context.Context,
	in service.DNSRecordsDeleteInput,
) (*service.DNSRecordsDeleteOutput, error) {
	path := "/dns-records/" + url.PathEscape(in.PayloadName) + "/" + strconv.FormatInt(in.Index, 10)

	var out service.DNSRecordsDeleteOutput
	if err := s.do(ctx, http.MethodDelete, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
