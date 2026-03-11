package clob

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/star-gazer111/poly-go-clob-client/types"
)

func TestNewPublicClient_BaseURLValidation(t *testing.T) {
	_, err := NewPublicClient("")
	if err == nil {
		t.Fatal("expected error for empty baseURL")
	}
	var te *types.Error
	if !errors.As(err, &te) || te.Kind() != types.KindValidation {
		t.Fatalf("expected KindValidation error, got: %T %v", err, err)
	}

	_, err = NewPublicClient("not a url")
	if err == nil {
		t.Fatal("expected error for invalid url")
	}
	if !errors.As(err, &te) || te.Kind() != types.KindValidation {
		t.Fatalf("expected KindValidation error, got: %T %v", err, err)
	}
}

func TestPublicClient_Ping_OK_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			t.Fatalf("expected /, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true,"message":"pong"}`))
	}))
	defer srv.Close()

	c, err := NewPublicClient(srv.URL) // http ok for tests
	if err != nil {
		t.Fatalf("NewPublicClient err: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out, err := c.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping err: %v", err)
	}
	if out == nil || !out.OK || out.Message != "pong" {
		t.Fatalf("unexpected ping response: %+v", out)
	}
}

func TestPublicClient_Ping_Non2xx_ReturnsTypedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`rate limited`))
	}))
	defer srv.Close()

	c, err := NewPublicClient(srv.URL)
	if err != nil {
		t.Fatalf("NewPublicClient err: %v", err)
	}

	_, err = c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	// must be a typed Status error
	var st *types.Status
	if !errors.As(err, &st) {
		t.Fatalf("expected errors.As to find *types.Status, got %T %v", err, err)
	}
	if st.StatusCode != 429 {
		t.Fatalf("expected 429, got %d", st.StatusCode)
	}
	if st.Path != "/" {
		t.Fatalf("expected path /, got %s", st.Path)
	}

	// top level kind should be KindStatus
	var top *types.Error
	if !errors.As(err, &top) {
		t.Fatalf("expected errors.As to find *types.Error, got %T %v", err, err)
	}
	if top.Kind() != types.KindStatus {
		t.Fatalf("expected kind %s, got %s", types.KindStatus, top.Kind())
	}
}

func TestPublicClient_GetMarketTradesEvents_UsesDataAPI(t *testing.T) {
	var gotQuery url.Values

	dataSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trades" {
			t.Fatalf("expected /trades, got %s", r.URL.Path)
		}
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
			"proxyWallet":"0x56687bf447db6ffa42ffe2204a05edaa20f55839",
			"side":"BUY",
			"asset":"123",
			"conditionId":"0xdd22472e552920b8438158ea7238bfadfa4f736aa4cee91a6b86c39ead110917",
			"size":123.45,
			"price":0.67,
			"timestamp":1710000000,
			"title":"Example",
			"slug":"example-market",
			"icon":"https://example.com/icon.png",
			"eventSlug":"example-event",
			"outcome":"Yes",
			"outcomeIndex":0,
			"name":"Trader",
			"pseudonym":"Alias",
			"bio":"",
			"profileImage":"",
			"profileImageOptimized":"",
			"transactionHash":"0xabc"
		}]`))
	}))
	defer dataSrv.Close()

	c, err := NewPublicClient("https://clob.polymarket.com", WithDataAPIURL(dataSrv.URL))
	if err != nil {
		t.Fatalf("NewPublicClient err: %v", err)
	}

	takerOnly := false
	resp, err := c.GetMarketTradesEvents(context.Background(), &types.GetMarketTradesEventsRequest{
		ConditionID: "0xdd22472e552920b8438158ea7238bfadfa4f736aa4cee91a6b86c39ead110917",
		Limit:       25,
		Offset:      10,
		TakerOnly:   &takerOnly,
		Side:        "BUY",
	})
	if err != nil {
		t.Fatalf("GetMarketTradesEvents err: %v", err)
	}

	if gotQuery.Get("market") != "0xdd22472e552920b8438158ea7238bfadfa4f736aa4cee91a6b86c39ead110917" {
		t.Fatalf("expected market query param, got %q", gotQuery.Get("market"))
	}
	if gotQuery.Get("limit") != "25" {
		t.Fatalf("expected limit=25, got %q", gotQuery.Get("limit"))
	}
	if gotQuery.Get("offset") != "10" {
		t.Fatalf("expected offset=10, got %q", gotQuery.Get("offset"))
	}
	if gotQuery.Get("takerOnly") != "false" {
		t.Fatalf("expected takerOnly=false, got %q", gotQuery.Get("takerOnly"))
	}
	if gotQuery.Get("side") != "BUY" {
		t.Fatalf("expected side=BUY, got %q", gotQuery.Get("side"))
	}

	if len(resp) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(resp))
	}
	if resp[0].ConditionID != "0xdd22472e552920b8438158ea7238bfadfa4f736aa4cee91a6b86c39ead110917" {
		t.Fatalf("unexpected condition id: %s", resp[0].ConditionID)
	}
	if !resp[0].Price.Equal(decimal.RequireFromString("0.67")) {
		t.Fatalf("unexpected price: %s", resp[0].Price)
	}
}
