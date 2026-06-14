package remotesvc

import (
	"context"
	"net/http"
	"strconv"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) AuditRecordsGet(
	ctx context.Context,
	in service.AuditRecordsGetInput,
) (*service.AuditRecordsGetOutput, error) {
	var out service.AuditRecordsGetOutput
	if err := s.do(ctx, http.MethodGet, "/audit-records/"+strconv.FormatInt(in.ID, 10), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
