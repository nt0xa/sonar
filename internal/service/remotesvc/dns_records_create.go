package remotesvc

import (
	"context"
	"net/http"

	"github.com/nt0xa/sonar/internal/modules/api2/apimodels"
	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) DNSRecordsCreate(
	ctx context.Context,
	in service.DNSRecordsCreateInput,
) (*service.DNSRecordsCreateOutput, error) {
	req := apimodels.DNSRecordsCreateRequest{
		PayloadName: in.PayloadName,
		Name:        in.Name,
		TTL:         in.TTL,
		Type:        in.Type,
		Values:      in.Values,
		Strategy:    in.Strategy,
	}

	var out service.DNSRecordsCreateOutput
	if err := s.do(ctx, http.MethodPost, "/dns-records", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
