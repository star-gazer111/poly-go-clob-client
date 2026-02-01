package clob

import (
	"context"
	"net/http"

	"github.com/star-gazer111/poly-go-clob-client/internal/transport"
)

type PublicClient struct {
	baseURL   string
	transport *transport.Transport
}

type PublicClientOption func(*PublicClient)

func NewPublicClient(baseURL string, opts ...PublicClientOption) *PublicClient {
	c := &PublicClient{
		baseURL:   baseURL,
		transport: transport.NewTransport(http.DefaultClient, transport.DefaultPolicy()),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithHTTPClient(hc *http.Client) PublicClientOption {
	return func(c *PublicClient) {
		c.transport = transport.NewTransport(hc, transport.DefaultPolicy())
	}
}

func WithTransport(t *transport.Transport) PublicClientOption {
	return func(c *PublicClient) {
		c.transport = t
	}
}

// Example public method placeholder
func (c *PublicClient) Ping(ctx context.Context) error {
	_, err := c.transport.DoJSON(ctx, http.MethodGet, c.baseURL+"/ping", nil, nil)
	return err
}
