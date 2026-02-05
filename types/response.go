package types

import "time"

// OrderSummary represents a single price level in the order book.
type OrderSummary struct {
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
}

// OrderBookSummaryResponse represents the response from the order book summary endpoint.
type OrderBookSummaryResponse struct {
	Market         string         `json:"market"`
	AssetID        string         `json:"asset_id"`
	Timestamp      time.Time      `json:"timestamp"`
	Hash           *string        `json:"hash,omitempty"`
	Bids           []OrderSummary `json:"bids"`
	Asks           []OrderSummary `json:"asks"`
	MinOrderSize   float64        `json:"min_order_size"`
	NegRisk        bool           `json:"neg_risk"`
	TickSize       float64        `json:"tick_size"`
	LastTradePrice *float64       `json:"last_trade_price,omitempty"`
}
