package cache

import (
	"context"
	"sync"

	"github.com/nt0xa/sonar/internal/database"
)

type Cache interface {
	SubdomainExists(subdomain string) bool
}

type cache struct {
	subdomains sync.Map
}

func New(ctx context.Context, db *database.DB) (Cache, error) {
	c := &cache{
		subdomains: sync.Map{},
	}

	if err := c.loadSubdomains(ctx, db); err != nil {
		return nil, err
	}

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

func (c *cache) SubdomainExists(subdomain string) bool {
	_, ok := c.subdomains.Load(subdomain)
	return ok
}
