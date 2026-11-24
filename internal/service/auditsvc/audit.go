package auditsvc

import (
	"context"
	"encoding/json"
	"fmt"
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

// Record is a single audit entry; Resource is stored as JSON.
type Record struct {
	Action        database.AuditRecordActionType
	ResourceType  database.AuditRecordResourceType
	Source        database.AuditRecordSourceType
	ActorID       *int64
	ActorName     string
	ActorMetadata database.AuditActorMetadata
	Resource      any
}

// WriteRecord synchronously stores r in the audit log.
func WriteRecord(ctx context.Context, db *database.DB, r Record) error {
	raw, err := json.Marshal(r.Resource)
	if err != nil {
		return fmt.Errorf("marshal resource: %w", err)
	}

	meta := r.ActorMetadata
	if meta == nil {
		meta = database.AuditActorMetadata{}
	}

	_, err = db.AuditRecordsCreate(ctx, database.AuditRecordsCreateParams{
		Action:        r.Action,
		ResourceType:  r.ResourceType,
		Source:        r.Source,
		ActorID:       r.ActorID,
		ActorName:     r.ActorName,
		ActorMetadata: meta,
		Resource:      raw,
	})
	return err
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
	c, ok := service.CallerFrom(ctx)
	if !ok {
		s.log.Warn("skip audit: no actor in context", "resourceType", resourceType, "action", action)
		return
	}

	ctx = context.WithoutCancel(ctx)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		if err := WriteRecord(ctx, s.db, Record{
			Action:        action,
			ResourceType:  resourceType,
			Source:        database.AuditRecordSourceType(c.Source),
			ActorID:       &c.UserID,
			ActorName:     c.UserName,
			ActorMetadata: actorMetadata(c),
			Resource:      resource,
		}); err != nil {
			s.log.Warn("failed to write audit record", "err", err, "resourceType", resourceType, "action", action)
		}
	}()
}

// actorMetadata returns source-specific identifiers for the actor.
func actorMetadata(c service.Caller) database.AuditActorMetadata {
	meta := database.AuditActorMetadata{}

	switch c.Source {
	case service.AuditSourceApi:
		if c.RemoteIP != nil {
			meta["remoteIp"] = *c.RemoteIP
		}
	case service.AuditSourceTelegram:
		if c.TelegramID != nil {
			meta["telegramId"] = *c.TelegramID
		}
	case service.AuditSourceLark:
		if c.LarkID != nil {
			meta["larkId"] = *c.LarkID
		}
	case service.AuditSourceSlack:
		if c.SlackID != nil {
			meta["slackId"] = *c.SlackID
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
