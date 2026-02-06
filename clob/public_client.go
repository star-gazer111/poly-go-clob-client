package clob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/star-gazer111/poly-go-clob-client/internal/transport"
	"github.com/star-gazer111/poly-go-clob-client/types"
)

type PublicClient struct {
	baseURL      *url.URL
	transport    *transport.Transport
	requireHTTPS bool
}

// PublicClientOption configures the PublicClient
type PublicClientOption func(*PublicClient)

// WithTransport sets a custom transport
func WithTransport(t *transport.Transport) PublicClientOption {
	return func(c *PublicClient) {
		if t != nil {
			c.transport = t
		}
	}
}

// WithHTTPClient sets a custom http.Client while preserving default transport policy
func WithHTTPClient(hc *http.Client) PublicClientOption {
	return func(c *PublicClient) {
		if hc != nil {
			c.transport = transport.NewTransport(hc, transport.DefaultPolicy())
		}
	}
}

// WithRequireHTTPS enforces https scheme for baseURL.
func WithRequireHTTPS(require bool) PublicClientOption {
	return func(c *PublicClient) {
		c.requireHTTPS = require
	}
}

// NewPublicClient constructs a public-only client
//
// - baseURL must be a valid URL (non-empty)
// - https is recommended we can optionally enforce via WithRequireHTTPS(true)
func NewPublicClient(baseURL string, opts ...PublicClientOption) (*PublicClient, error) {
	c := &PublicClient{
		transport:    transport.NewTransport(http.DefaultClient, transport.DefaultPolicy()),
		requireHTTPS: false,
	}

	for _, opt := range opts {
		opt(c)
	}

	u, err := validateBaseURL(baseURL, c.requireHTTPS)
	if err != nil {
		return nil, err
	}
	c.baseURL = u
	return c, nil
}

func validateBaseURL(baseURL string, requireHTTPS bool) (*url.URL, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil, types.ValidationErr("baseURL must be non-empty")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, types.WithSource(types.KindValidation, err)
	}

	// must have scheme + host
	if u.Scheme == "" || u.Host == "" {
		return nil, types.ValidationErr("baseURL must include scheme and host (e.g., https://example.com)")
	}

	// https. Allow http for local/testing
	if requireHTTPS && strings.ToLower(u.Scheme) != "https" {
		return nil, types.ValidationErr("baseURL must use https scheme")
	}

	// we normalize: remove trailing slash path unless user has custom base path
	// and we keep any non-root path if present
	if u.Path == "/" {
		u.Path = ""
	}
	u.RawQuery = ""
	u.Fragment = ""

	return u, nil
}

func (c *PublicClient) endpoint(path string) string {
	// path should start with "/" for correctness
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// preserve baseURL path prefix if present
	base := *c.baseURL
	base.Path = strings.TrimRight(base.Path, "/") + path
	return base.String()
}

// PingResponse is a minimal response model - kept it intentionally small for v0.1
type PingResponse struct {
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
	OK      bool   `json:"ok,omitempty"`
	Raw     string `json:"-"`
}

// ping calls a simple public endpoint (e.g. /ping) and returns a parsed response
// if the server returns JSON we parse it otherwise we return raw string
//
// this is meant as a boring smoke-test endpoint for transport + error typing
func (c *PublicClient) Ping(ctx context.Context) (*PingResponse, error) {
	b, err := c.transport.DoJSON(ctx, http.MethodGet, c.endpoint("/ping"), nil, nil)
	if err != nil {
		return nil, err
	}

	resp := &PingResponse{}
	// try JSON first
	if json.Unmarshal(b, resp) == nil {
		return resp, nil
	}

	// fallback: treat as plain text
	resp.Raw = strings.TrimSpace(string(b))
	resp.OK = resp.Raw != ""
	return resp, nil
}

func (c *PublicClient) OrderBook(ctx context.Context, req *types.OrderBookSummaryRequest) (*types.OrderBookSummaryResponse, error) {
	q := url.Values{}
	q.Add("token_id", fmt.Sprintf("%d", req.TokenId))

	// Side is int, so we send as "0" (Buy) or "1" (Sell)

	q.Add("side", fmt.Sprintf("%d", req.Side))

	// Construct URL with query params
	u := c.endpoint("/book")
	if len(q) > 0 {
		u = u + "?" + q.Encode()
	}

	b, err := c.transport.DoJSON(ctx, http.MethodGet, u, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp types.OrderBookSummaryResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *PublicClient) OrderBooks(ctx context.Context, req []types.OrderBookSummaryRequest) ([]*types.OrderBookSummaryResponse, error) {
	u := c.endpoint("/books")

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	b, err := c.transport.DoJSON(ctx, http.MethodPost, u, nil, body)

	if err != nil {
		return nil, err
	}

	var resp []*types.OrderBookSummaryResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *PublicClient) Market(ctx context.Context, req *types.MarketRequest) (*types.MarketResponse, error) {
	u := c.endpoint(fmt.Sprintf("/markets/%s", req.ConditionID))

	b, err := c.transport.DoJSON(ctx, http.MethodGet, u, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp types.MarketResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Markets fetches a paginated list of all markets.
// Pass an empty string for nextCursor to get the first page.
func (c *PublicClient) Markets(ctx context.Context, nextCursor string) (*types.MarketsPage, error) {
	u := c.endpoint("/markets")
	if nextCursor != "" {
		u = u + "?next_cursor=" + url.QueryEscape(nextCursor)
	}

	b, err := c.transport.DoJSON(ctx, http.MethodGet, u, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp types.MarketsPage
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// SimplifiedMarkets fetches a paginated list of simplified markets.
// Pass an empty string for nextCursor to get the first page.
func (c *PublicClient) SimplifiedMarkets(ctx context.Context, nextCursor string) (*types.SimplifiedMarketsPage, error) {
	u := c.endpoint("/simplified-markets")
	if nextCursor != "" {
		u = u + "?next_cursor=" + url.QueryEscape(nextCursor)
	}

	b, err := c.transport.DoJSON(ctx, http.MethodGet, u, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp types.SimplifiedMarketsPage
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// SamplingMarkets fetches a paginated list of sampling markets.
// Pass an empty string for nextCursor to get the first page.
func (c *PublicClient) SamplingMarkets(ctx context.Context, nextCursor string) (*types.MarketsPage, error) {
	u := c.endpoint("/sampling-markets")
	if nextCursor != "" {
		u = u + "?next_cursor=" + url.QueryEscape(nextCursor)
	}

	b, err := c.transport.DoJSON(ctx, http.MethodGet, u, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp types.MarketsPage
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
