package main

import (
	"context"
	"net/url"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func init() {
	validation.ErrorTag = "mapstructure"
}

type Config struct {
	Context Context           `mapstructure:"context"`
	Servers map[string]Server `mapstructure:"servers"`
}

func (c Config) ValidateWithContext(ctx context.Context) error {

	servers := make([]interface{}, 0)

	for s := range c.Servers {
		servers = append(servers, s)
	}

	ctx = context.WithValue(ctx, "servers", servers)

	return validation.ValidateStructWithContext(ctx, &c,
		validation.Field(&c.Context),
		validation.Field(&c.Servers, validation.Length(1, 0)),
	)
}

func (c *Config) Server() *Server {
	srv, ok := c.Servers[c.Context.Server]
	if !ok {
		return nil
	}
	return &srv
}

type Context struct {
	Server string `mapstructure:"server"`
}

func (c Context) ValidateWithContext(ctx context.Context) error {
	servers, ok := ctx.Value("servers").([]interface{})
	if !ok {
		panic(`fail to find "servers" key in context`)
	}

	return validation.ValidateStructWithContext(ctx, &c,
		validation.Field(&c.Server, validation.Required, validation.In(servers...)),
	)
}

type Server struct {
	Token    string  `mapstructure:"token"`
	URL      string  `mapstructure:"url"`
	Proxy    *string `mapstructure:"proxy"`
	Insecure bool    `mapstructure:"insecure"`
}

func (c Server) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &c,
		validation.Field(&c.Token,
			validation.Required,
			validation.Match(regexp.MustCompile("[a-f0-9]{32}")),
		),
		validation.Field(&c.URL, validation.Required, is.URL),
		validation.Field(&c.Proxy, is.URL),
	)
}

func (c *Server) BaseURL() *url.URL {
	u, err := url.Parse(c.URL)
	if err != nil {
		// Already passed validation, must be valid URL.
		panic(err)
	}

	return u
}

func (c *Server) ProxyURL() *url.URL {
	if c.Proxy == nil {
		return nil
	}

	u, err := url.Parse(*c.Proxy)
	if err != nil {
		// Already passed validation, must be valid URL.
		panic(err)
	}

	return u
}
