package main

import (
	"net/url"
	"regexp"

	"github.com/nt0xa/sonar/pkg/valid"
)

type Config struct {
	Context Context           `mapstructure:"context"`
	Servers map[string]Server `mapstructure:"servers"`
}

func (c Config) Validate() valid.Problems {
	servers := make([]string, 0, len(c.Servers))
	for name := range c.Servers {
		servers = append(servers, name)
	}

	fields := []valid.Validatable{
		valid.Slice("servers", servers, valid.NotEmpty),
		valid.String("context.server", c.Context.Server, valid.Required, valid.In(servers...)),
	}
	for name, srv := range c.Servers {
		fields = append(fields, valid.Struct("servers."+name, srv))
	}

	return valid.Validate(fields...)
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

type Server struct {
	Token    string  `mapstructure:"token"`
	URL      string  `mapstructure:"url"`
	Proxy    *string `mapstructure:"proxy"`
	Insecure bool    `mapstructure:"insecure"`
}

var tokenRe = regexp.MustCompile("[a-f0-9]{32}")

func (c Server) Validate() valid.Problems {
	return valid.Validate(
		valid.String("token", c.Token, valid.Required, valid.Match(tokenRe, "invalid token")),
		valid.String("url", c.URL, valid.Required, valid.URL),
		valid.OptionalString("proxy", c.Proxy, valid.URL),
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
