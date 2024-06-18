package apiclient

import (
	"crypto/tls"
	"fmt"

	"github.com/go-resty/resty/v2"

	"github.com/nt0xa/sonar/internal/actions"
)

type Client struct {
	client *resty.Client
}

var _ actions.Actions = &Client{}

func New(url string, token string, insecure bool, proxy *string) *Client {
	c := resty.New().
		SetHostURL(url).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		SetHeader("Content-Type", "application/json")

	if proxy != nil {
		c.SetProxy(*proxy)
	}

	if insecure {
		c.SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
		})
	}

	return &Client{c}
}
