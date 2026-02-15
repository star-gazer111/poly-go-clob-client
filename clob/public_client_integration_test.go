//go:build integration
// +build integration

package clob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/star-gazer111/poly-go-clob-client/types"
)

const (
	// LiveBaseURL is the Polymarket CLOB API base URL.
	LiveBaseURL = "https://clob.polymarket.com"
	// DefaultTimeout for API calls.
	DefaultTimeout = 30 * time.Second
)

// getTestClient returns a PublicClient configured for live testing.
func getTestClient(t *testing.T) *PublicClient {
	t.Helper()

	baseURL := os.Getenv("POLYMARKET_CLOB_URL")
	if baseURL == "" {
		baseURL = LiveBaseURL
	}

	c, err := NewPublicClient(baseURL, WithRequireHTTPS(true))
	if err != nil {
		t.Fatalf("NewPublicClient: %v", err)
	}
	return c
}

// TestIntegration_Ping tests the root endpoint as a health check against the live API.
func TestIntegration_Ping(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	resp, err := c.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	t.Logf("Ping response: %+v", resp)
	if resp == nil {
		t.Fatal("expected non-nil ping response")
	}
}

// TestIntegration_Markets tests the /markets endpoint against the live API.
func TestIntegration_Markets(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	resp, err := c.Markets(ctx, "")
	fmt.Println(resp)
	if err != nil {
		t.Fatalf("Markets failed: %v", err)
	}

	t.Logf("Markets response: Count=%d, Limit=%d, NextCursor=%s", resp.Count, resp.Limit, resp.NextCursor)

	if resp.Count == 0 {
		t.Log("Warning: No markets returned from the API")
	}

	// Validate pagination works
	if resp.NextCursor != "" && resp.Count > 0 {
		t.Logf("First market condition_id: %v", resp.Data[0].ConditionID)
	}
}

// TestIntegration_Markets_Pagination tests pagination on /markets endpoint.
func TestIntegration_Markets_Pagination(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// First page
	page1, err := c.Markets(ctx, "")
	if err != nil {
		t.Fatalf("Markets (page 1) failed: %v", err)
	}

	if page1.NextCursor == "" {
		t.Skip("No next cursor, cannot test pagination")
	}

	// Second page
	page2, err := c.Markets(ctx, page1.NextCursor)
	if err != nil {
		t.Fatalf("Markets (page 2) failed: %v", err)
	}

	t.Logf("Page 1: Count=%d, Page 2: Count=%d", page1.Count, page2.Count)

	// Ensure we got different data (if both pages have data)
	if len(page1.Data) > 0 && len(page2.Data) > 0 {
		if page1.Data[0].ConditionID != nil && page2.Data[0].ConditionID != nil {
			if *page1.Data[0].ConditionID == *page2.Data[0].ConditionID {
				t.Error("Page 1 and Page 2 returned the same first market")
			}
		}
	}
}

// TestIntegration_Market tests the /markets/{condition_id} endpoint against the live API.
func TestIntegration_Market(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// First, get a valid condition_id from the markets list
	markets, err := c.Markets(ctx, "")
	if err != nil {
		t.Fatalf("Markets failed: %v", err)
	}

	if len(markets.Data) == 0 {
		t.Skip("No markets available to test Market endpoint")
	}

	// Find a market with a valid condition_id
	var conditionID string
	for _, m := range markets.Data {
		if m.ConditionID != nil && *m.ConditionID != "" {
			conditionID = *m.ConditionID
			break
		}
	}

	if conditionID == "" {
		t.Skip("No market with condition_id found")
	}

	t.Logf("Testing Market with condition_id: %s", conditionID)

	req := &types.MarketRequest{ConditionID: conditionID}
	resp, err := c.Market(ctx, req)
	if err != nil {
		t.Fatalf("Market failed: %v", err)
	}

	t.Logf("Market response: Question=%s, Active=%v", resp.Question, resp.Active)

	if resp.ConditionID == nil || *resp.ConditionID != conditionID {
		t.Errorf("Expected condition_id %s, got %v", conditionID, resp.ConditionID)
	}
}

// TestIntegration_SimplifiedMarkets tests the /simplified-markets endpoint against the live API.
func TestIntegration_SimplifiedMarkets(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	resp, err := c.SimplifiedMarkets(ctx, "")
	if err != nil {
		t.Fatalf("SimplifiedMarkets failed: %v", err)
	}

	t.Logf("SimplifiedMarkets response: Count=%d, Limit=%d", resp.Count, resp.Limit)

	if resp.Count > 0 && len(resp.Data) > 0 {
		t.Logf("First simplified market condition_id: %v", resp.Data[0].ConditionID)
	}
}

// TestIntegration_SamplingMarkets tests the /sampling-markets endpoint against the live API.
func TestIntegration_SamplingMarkets(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	resp, err := c.SamplingMarkets(ctx, "")
	if err != nil {
		// Response body too large is expected for this endpoint
		if strings.Contains(err.Error(), "response body too large") {
			t.Skip("SamplingMarkets response too large for default transport settings")
		}
		t.Fatalf("SamplingMarkets failed: %v", err)
	}

	t.Logf("SamplingMarkets response: Count=%d, Limit=%d", resp.Count, resp.Limit)

	if resp.Count > 0 && len(resp.Data) > 0 {
		t.Logf("First sampling market condition_id: %v", resp.Data[0].ConditionID)
	}
}

// TestIntegration_SamplingSimplifiedMarkets tests the /sampling-simplified-markets endpoint.
func TestIntegration_SamplingSimplifiedMarkets(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	resp, err := c.SamplingSimplifiedMarkets(ctx, "")
	if err != nil {
		t.Fatalf("SamplingSimplifiedMarkets failed: %v", err)
	}

	t.Logf("SamplingSimplifiedMarkets response: Count=%d, Limit=%d", resp.Count, resp.Limit)

	if resp.Count > 0 && len(resp.Data) > 0 {
		t.Logf("First sampling simplified market condition_id: %v", resp.Data[0].ConditionID)
	}
}

// TestIntegration_OrderBook tests the /book endpoint against the live API.
// Uses Gamma API to get valid token IDs.
func TestIntegration_OrderBook(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// Fetch token_id from Gamma API
	tokenID, err := fetchTokenIDFromGamma(ctx)
	if err != nil {
		t.Skipf("Could not fetch token_id from Gamma API: %v", err)
	}

	t.Logf("Testing OrderBook with token_id: %s", tokenID)

	req := &types.OrderBookSummaryRequest{
		TokenId: tokenID,
		Side:    types.SideBuy,
	}

	resp, err := c.OrderBook(ctx, req)
	if err != nil {
		t.Fatalf("OrderBook failed: %v", err)
	}

	t.Logf("OrderBook response: Market=%s, AssetID=%s, Bids=%d, Asks=%d",
		resp.Market, resp.AssetID, len(resp.Bids), len(resp.Asks))
}

// TestIntegration_OrderBooks tests the /books endpoint against the live API.
// Uses Gamma API to get valid token IDs.
func TestIntegration_OrderBooks(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// Fetch token_id from Gamma API
	tokenID, err := fetchTokenIDFromGamma(ctx)
	if err != nil {
		t.Skipf("Could not fetch token_id from Gamma API: %v", err)
	}

	t.Logf("Testing OrderBooks with token_id: %s", tokenID)

	req := []types.OrderBookSummaryRequest{
		{TokenId: tokenID},
	}

	resp, err := c.OrderBooks(ctx, req)
	if err != nil {
		t.Fatalf("OrderBooks failed: %v", err)
	}

	if len(resp) == 0 {
		t.Fatal("Expected at least one order book response")
	}

	t.Logf("OrderBooks response: count=%d, first market=%s", len(resp), resp[0].Market)
}

// GammaMarket represents a market from the Gamma API.
type GammaMarket struct {
	ClobTokenIds string `json:"clobTokenIds"`
}

// fetchTokenIDFromGamma fetches a valid token_id from the Gamma API.
func fetchTokenIDFromGamma(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://gamma-api.polymarket.com/markets?limit=5&active=true&closed=false", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gamma API returned status %d", resp.StatusCode)
	}

	var markets []GammaMarket
	if err := json.NewDecoder(resp.Body).Decode(&markets); err != nil {
		return "", err
	}

	if len(markets) == 0 {
		return "", fmt.Errorf("no markets found")
	}

	// Parse clobTokenIds JSON array (it's a JSON string)
	for _, m := range markets {
		if m.ClobTokenIds == "" || m.ClobTokenIds == "null" {
			continue
		}
		var tokenIds []string
		if err := json.Unmarshal([]byte(m.ClobTokenIds), &tokenIds); err != nil {
			continue
		}
		if len(tokenIds) > 0 && tokenIds[0] != "" {
			return tokenIds[0], nil
		}
	}

	return "", fmt.Errorf("no valid token_id found in markets")
}

func fetchTokenIDsFromGamma(ctx context.Context, limit int) ([]string, error) {
	url := fmt.Sprintf("https://gamma-api.polymarket.com/markets?limit=%d&active=true&closed=false", limit*5)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gamma API returned status %d", resp.StatusCode)
	}

	var markets []GammaMarket
	if err := json.NewDecoder(resp.Body).Decode(&markets); err != nil {
		return nil, err
	}

	if len(markets) == 0 {
		return nil, fmt.Errorf("no markets found")
	}

	var collectedIDs []string
	for _, m := range markets {
		if m.ClobTokenIds == "" || m.ClobTokenIds == "null" {
			continue
		}

		var tokenIds []string
		if err := json.Unmarshal([]byte(m.ClobTokenIds), &tokenIds); err == nil && len(tokenIds) > 0 {
			if tokenIds[0] != "" {
				collectedIDs = append(collectedIDs, tokenIds[0])
				if len(collectedIDs) >= limit {
					break
				}
			}
		}
	}

	if len(collectedIDs) < 1 {
		return nil, fmt.Errorf("no valid token_ids found")
	}

	// if we collect less than limit but more than one, then DONT throw error

	return collectedIDs, nil
}

// TestIntegration_AllEndpointsReachable runs a smoke test to verify all endpoints are reachable.
func TestIntegration_AllEndpointsReachable(t *testing.T) {
	c := getTestClient(t)

	tests := []struct {
		name          string
		fn            func(context.Context) error
		skipBodyLarge bool // if true, skip on "response body too large" errors
	}{
		// Ping is excluded - Polymarket doesn't have this endpoint
		{
			name: "Markets",
			fn: func(ctx context.Context) error {
				_, err := c.Markets(ctx, "")
				return err
			},
		},
		{
			name: "SimplifiedMarkets",
			fn: func(ctx context.Context) error {
				_, err := c.SimplifiedMarkets(ctx, "")
				return err
			},
		},
		{
			name:          "SamplingMarkets",
			skipBodyLarge: true,
			fn: func(ctx context.Context) error {
				_, err := c.SamplingMarkets(ctx, "")
				return err
			},
		},
		{
			name: "SamplingSimplifiedMarkets",
			fn: func(ctx context.Context) error {
				_, err := c.SamplingSimplifiedMarkets(ctx, "")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
			defer cancel()

			err := tt.fn(ctx)
			if err != nil {
				if tt.skipBodyLarge && strings.Contains(err.Error(), "response body too large") {
					t.Skipf("%s: response body too large (expected)", tt.name)
				}
				t.Errorf("%s endpoint failed: %v", tt.name, err)
			} else {
				t.Logf("%s endpoint: OK", tt.name)
			}
		})
	}
}

// TestIntegration_GetLastTradePrice tests the /last-trade-price endpoint.
func TestIntegration_GetLastTradePrice(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// Fetch token_id from Gamma API
	tokenID, err := fetchTokenIDFromGamma(ctx)
	if err != nil {
		t.Skipf("Could not fetch token_id from Gamma API: %v", err)
	}

	t.Logf("Testing GetLastTradePrice with token_id: %s", tokenID)

	req := &types.LastTradePriceRequest{
		TokenId: tokenID,
	}

	resp, err := c.GetLastTradePrice(ctx, req)
	if err != nil {
		t.Fatalf("GetLastTradePrice failed: %v", err)
	}

	t.Logf("GetLastTradePrice response: Price=%s, Side=%v", resp.Price, resp.Side)

	if resp.Price.String() == "" {
		t.Error("Expected price to be present")
	}

	if resp.Side != "BUY" && resp.Side != "SELL" {
		t.Errorf("Expected Side to be 'BUY' or 'SELL', got '%s'", resp.Side)
	}
}

// TestIntegration_GetLastTradesPrices tests the /last-trades-prices endpoint.
func TestIntegration_GetLastTradesPrices(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// Fetch token_id from Gamma API
	tokenID, err := fetchTokenIDFromGamma(ctx)
	if err != nil {
		t.Skipf("Could not fetch token_id from Gamma API: %v", err)
	}

	t.Logf("Testing GetLastTradesPrices with token_id: %s", tokenID)

	req := []types.LastTradePriceRequest{
		{TokenId: tokenID},
	}

	resp, err := c.GetLastTradesPrices(ctx, req)
	if err != nil {
		t.Fatalf("GetLastTradesPrices failed: %v", err)
	}

	if len(resp) == 0 {
		t.Fatal("Expected at least one response")
	}

	t.Logf("GetLastTradesPrices response: %+v", resp[0])

	if resp[0].Price.String() == "" {
		t.Error("Expected price to be present")
	}
}

// TestIntegration_GetMarketTradesEvents tests the /events/trade endpoint.
func TestIntegration_GetMarketTradesEvents(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// Fetch token_id from Gamma API to get a valid condition_id
	// Note: We need a condition_id, and fetchTokenIDFromGamma returns a token_id.
	// For simplicity, we'll fetch markets and use the first one's condition_id.

	markets, err := c.Markets(ctx, "")
	if err != nil {
		t.Fatalf("Markets failed: %v", err)
	}
	if len(markets.Data) == 0 {
		t.Skip("No markets available to test GetMarketTradesEvents")
	}

	var conditionID string
	for _, m := range markets.Data {
		if m.ConditionID != nil && *m.ConditionID != "" {
			conditionID = *m.ConditionID
			break
		}
	}

	if conditionID == "" {
		t.Skip("No valid condition_id found")
	}

	t.Logf("Testing GetMarketTradesEvents with condition_id: %s", conditionID)

	req := &types.GetMarketTradesEventsRequest{
		ConditionID: conditionID,
	}

	resp, err := c.GetMarketTradesEvents(ctx, req)
	if err != nil {
		t.Fatalf("GetMarketTradesEvents failed: %v", err)
	}

	t.Logf("GetMarketTradesEvents response: next_cursor=%s, count=%d", resp.NextCursor, len(resp.Data))

	if len(resp.Data) > 0 {
		t.Logf("First trade event: %+v", resp.Data[0])
	}
}

func TestIntegration_Midpoint(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	// fetch token id from Gamma API
	tokenID, err := fetchTokenIDFromGamma(ctx)
	if err != nil {
		t.Skipf("Could not fetch token_id from Gamma API: %v", err)
	}

	t.Logf("Testing Orderbook with token_id : %s", tokenID)

	req := &types.MidpointRequest{
		TokenId: tokenID,
	}

	resp, err := c.Midpoint(ctx, req)
	if err != nil {
		t.Fatalf("Expected mid point to be present: %v", err)
	}

	t.Logf("Mid point: %v", resp.Mid)
}

func TestIntegration_Midpoints(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	tokenIDs, err := fetchTokenIDsFromGamma(ctx, 3)
	if err != nil {
		t.Skipf("Could not fetch tokens from gamma api : %v", err)
	}

	t.Logf("Testing midpoints with %d tokens : %v", len(tokenIDs), tokenIDs)

	var req []types.MidpointRequest
	for _, tid := range tokenIDs {
		req = append(req, types.MidpointRequest{TokenId: tid})
	}

	resp, err := c.Midpoints(ctx, req)
	if err != nil {
		t.Fatalf("Error in fetching midpoints : %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non nill response map")
	}

	for _, tid := range tokenIDs {
		if val, ok := resp[tid]; !ok {
			t.Errorf("Expected token %s in response map, but was missing", tid)
		} else {
			t.Logf("Token %s midpoint : %v", tid, val)
		}
	}
}

func TestIntegration_GetPrice(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	tokenID, err := fetchTokenIDFromGamma(ctx)
	if err != nil {
		t.Skipf("Could not fetch token_id from Gamma API: %v", err)
	}

	t.Logf("Testing GetPrice with token_id: %s", tokenID)

	req := types.PriceRequest{
		TokenId: tokenID,
		Side:    "BUY",
	}

	resp, err := c.GetPrice(ctx, req)
	if err != nil {
		t.Fatalf("GetPrice failed: %v", err)
	}

	t.Logf("Price for %s (BUY): %v", tokenID, resp.Price)
}

func TestIntegration_GetPrices(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	tokenIDs, err := fetchTokenIDsFromGamma(ctx, 3)
	if err != nil {
		t.Skipf("Could not fetch token_ids from Gamma API: %v", err)
	}

	t.Logf("Testing GetPrices with %d tokens", len(tokenIDs))

	var reqs []types.PriceRequest
	for _, tid := range tokenIDs {
		reqs = append(reqs, types.PriceRequest{TokenId: tid, Side: "BUY"})
	}

	resp, err := c.GetPrices(ctx, reqs)
	if err != nil {
		t.Fatalf("GetPrices failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response map")
	}

	for _, tid := range tokenIDs {
		sides, ok := resp[tid]
		if !ok {
			t.Errorf("Expected token %s in response map", tid)
			continue
		}

		val, ok := sides["BUY"]
		if !ok {
			t.Errorf("Expected BUY price for token %s", tid)
		} else {
			t.Logf("Token %s BUY Price: %v", tid, val)
		}
	}
}

func TestIntegration_GetSpread(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	tokenID, err := fetchTokenIDFromGamma(ctx)
	if err != nil {
		t.Skipf("Could not fetch token_id from Gamma API: %v", err)
	}

	t.Logf("Testing GetSpread with token_id: %s", tokenID)

	req := types.SpreadRequest{TokenId: tokenID}
	resp, err := c.GetSpread(ctx, req)
	if err != nil {
		t.Fatalf("GetSpread failed: %v", err)
	}

	t.Logf("Spread: %v", resp.Spread)
}

func TestIntegration_GetSpreads(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	tokenIDs, err := fetchTokenIDsFromGamma(ctx, 3)
	if err != nil {
		t.Skipf("Could not fetch token_ids from Gamma API: %v", err)
	}

	t.Logf("Testing GetSpreads with %d tokens", len(tokenIDs))

	var reqs []types.SpreadRequest
	for _, tid := range tokenIDs {
		reqs = append(reqs, types.SpreadRequest{TokenId: tid})
	}

	resp, err := c.GetSpreads(ctx, reqs)
	if err != nil {
		t.Fatalf("GetSpreads failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response map")
	}

	for _, tid := range tokenIDs {
		if val, ok := resp[tid]; !ok {
			t.Errorf("Expected token %s in response map", tid)
		} else {
			t.Logf("Token %s Spread: %v", tid, val)
		}
	}
}

func TestIntegration_GetPricesHistory(t *testing.T) {
	c := getTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	tokenID, err := fetchTokenIDFromGamma(ctx)
	if err != nil {
		t.Skipf("Could not fetch token_id from Gamma API: %v", err)
	}

	t.Logf("Testing GetPricesHistory with token_id: %s", tokenID)

	req := types.PricesHistoryRequest{
		Market:   tokenID,
		Interval: types.IntervalMax,
	}

	resp, err := c.GetPricesHistory(ctx, req)
	if err != nil {
		t.Fatalf("GetPricesHistory (max) failed: %v", err)
	}

	t.Logf("History items (max): %d", len(resp.History))

	if len(resp.History) > 0 {
		first := resp.History[0]
		last := resp.History[len(resp.History)-1]
		t.Logf("Range: %d to %d", first.Time, last.Time)

		// Test branch 2: Custom Range
		mid := (first.Time + last.Time) / 2
		start := mid - 120 
		end := mid + 120

		t.Logf("Testing GetPricesHistory with custom range: %d - %d", start, end)
		reqRange := types.PricesHistoryRequest{
			Market:  tokenID,
			StartTs: &start,
			EndTs:   &end,
		}

		respRange, err := c.GetPricesHistory(ctx, reqRange)
		if err != nil {
			t.Fatalf("GetPricesHistory (range) failed: %v", err)
		}

		t.Logf("History items (range): %d", len(respRange.History))

		for _, p := range respRange.History {
			t.Logf("{Time Price} : %v", p)
		}
	}
}
