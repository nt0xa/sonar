package cache

import (
	"context"
	"sync"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/database/models"
)

type Cache interface {
	SubdomainExists(subdomain string) bool
}

type cache struct {
	database.DefaultObserver
	subdomains sync.Map
}

func New(ctx context.Context, db *database.DB) (Cache, error) {
	c := &cache{
		subdomains: sync.Map{},
	}

	if err := c.loadSubdomains(ctx, db); err != nil {
		return nil, err
	}

	db.Observe(c)

	return c, nil
}

func (c *cache) loadSubdomains(ctx context.Context, db *database.DB) error {
	subs, err := db.PayloadsGetAllSubdomains(ctx)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		c.subdomains.Store(sub, struct{}{})
	}

	return nil
}

func (c *cache) PayloadCreated(p models.Payload) {
	c.subdomains.Store(p.Subdomain, struct{}{})
}

func (c *cache) PayloadDeleted(p models.Payload) {
	c.subdomains.Delete(p.Subdomain)
}

func (c *cache) SubdomainExists(subdomain string) bool {
	_, ok := c.subdomains.Load(subdomain)
	return ok
}
