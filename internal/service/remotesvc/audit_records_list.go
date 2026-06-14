package remotesvc

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/nt0xa/sonar/internal/service"
)

func (s *Service) AuditRecordsList(
	ctx context.Context,
	in service.AuditRecordsListInput,
) (service.AuditRecordsListOutput, error) {
	q := url.Values{}
	if in.ActorID != nil {
		q.Set("actorId", strconv.FormatInt(*in.ActorID, 10))
	}
	if in.ActorName != "" {
		q.Set("actorName", in.ActorName)
	}
	if in.ResourceType != "" {
		q.Set("resourceType", string(in.ResourceType))
	}
	if in.Action != "" {
		q.Set("action", string(in.Action))
	}
	if in.From != nil {
		q.Set("from", in.From.Format(time.RFC3339))
	}
	if in.To != nil {
		q.Set("to", in.To.Format(time.RFC3339))
	}
	if in.Page != 0 {
		q.Set("page", strconv.FormatUint(uint64(in.Page), 10))
	}
	if in.PerPage != 0 {
		q.Set("perPage", strconv.FormatUint(uint64(in.PerPage), 10))
	}

	var out service.AuditRecordsListOutput
	if err := s.do(ctx, http.MethodGet, withQuery("/audit-records", q), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
