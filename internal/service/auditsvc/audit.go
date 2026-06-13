package auditsvc

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// Service decorates another [service.ServerService] and writes an audit record
// after each successful mutating operation. Non-mutating methods (including the
// AuthContext* identity resolvers) pass through to the embedded service.
type Service struct {
	service.ServerService
	db  *database.DB
	log *slog.Logger
	wg  sync.WaitGroup
}

func New(svc service.ServerService, db *database.DB, log *slog.Logger) service.ServerService {
	return &Service{ServerService: svc, db: db, log: log}
}

var _ service.ServerService = (*Service)(nil)

// Wait blocks until all in-flight background audit writes have finished. It is
// intended for graceful shutdown and for deterministic tests.
func (s *Service) Wait() {
	s.wg.Wait()
}

// writeAudit records a single audit entry in the background. It is best-effort:
// the write runs in its own goroutine on a detached context so the caller is
// never blocked on it and a finished request can't cancel it, and any failure
// is logged rather than propagated.
func (s *Service) writeAudit(
	ctx context.Context,
	action database.AuditRecordActionType,
	resourceType database.AuditRecordResourceType,
	resource any,
) {
	id, ok := service.GetUserID(ctx)
	if !ok {
		s.log.Warn("skip audit: no actor in context", "resourceType", resourceType, "action", action)
		return
	}

	source := database.AuditRecordSourceTypeAPI
	if src, ok := service.GetSource(ctx); ok {
		source = database.AuditRecordSourceType(src)
	}

	ctx = context.WithoutCancel(ctx)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		raw, err := json.Marshal(resource)
		if err != nil {
			s.log.Warn("skip audit: failed to marshal resource", "err", err, "resourceType", resourceType)
			return
		}

		u, err := s.db.UsersGetByID(ctx, id)
		if err != nil {
			s.log.Warn("skip audit: failed to resolve actor", "err", err, "actorId", id)
			return
		}

		_, err = s.db.AuditRecordsCreate(ctx, database.AuditRecordsCreateParams{
			Action:        action,
			ResourceType:  resourceType,
			Source:        source,
			ActorID:       &u.ID,
			ActorName:     u.Name,
			ActorMetadata: actorMetadata(source, u),
			Resource:      raw,
		})
		if err != nil {
			s.log.Warn("failed to write audit record", "err", err, "resourceType", resourceType, "action", action)
		}
	}()
}

// actorMetadata returns source-specific identifiers for the actor.
func actorMetadata(source database.AuditRecordSourceType, u *database.User) database.AuditActorMetadata {
	meta := database.AuditActorMetadata{}

	switch source {
	case database.AuditRecordSourceTypeTelegram:
		if u.TelegramID != nil {
			meta["telegramId"] = *u.TelegramID
		}
	case database.AuditRecordSourceTypeLark:
		if u.LarkID != nil {
			meta["larkId"] = *u.LarkID
		}
	case database.AuditRecordSourceTypeSlack:
		if u.SlackID != nil {
			meta["slackId"] = *u.SlackID
		}
	}

	return meta
}

// maskAPIToken returns a copy of u with the API token cleared, so it is never
// recorded in the audit log.
func maskAPIToken(u *service.User) service.User {
	out := *u
	out.APIToken = nil
	return out
}
