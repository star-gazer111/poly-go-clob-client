package types

import (
	"time"

	"github.com/shopspring/decimal"
)

// Token represents a market token.
type Token struct {
	TokenID string `json:"token_id"`
	Outcome string `json:"outcome"`
	Price   string `json:"price,omitempty"`
	Winner  bool   `json:"winner"`
}

// Rewards represents market rewards configuration.
type Rewards struct {
	Rates     []RewardRate `json:"rates"`
	MinSize   string       `json:"min_size"`
	MaxSpread string       `json:"max_spread"`
}

// RewardRate represents a reward rate entry.
type RewardRate struct {
	AssetAddress string `json:"asset_address"`
	RewardsDaily string `json:"rewards_daily"`
}

// MarketResponse represents a market from the CLOB API.
type MarketResponse struct {
	EnableOrderBook         bool            `json:"enable_order_book"`
	Active                  bool            `json:"active"`
	Closed                  bool            `json:"closed"`
	Archived                bool            `json:"archived"`
	AcceptingOrders         bool            `json:"accepting_orders"`
	AcceptingOrderTimestamp *time.Time      `json:"accepting_order_timestamp,omitempty"`
	MinimumOrderSize        decimal.Decimal `json:"minimum_order_size"`
	MinimumTickSize         decimal.Decimal `json:"minimum_tick_size"`
	// ConditionID is the market condition ID (unique market identifier).
	ConditionID *string `json:"condition_id,omitempty"`
	// QuestionID is the CTF question ID.
	QuestionID    *string    `json:"question_id,omitempty"`
	Question      string     `json:"question"`
	Description   string     `json:"description"`
	MarketSlug    string     `json:"market_slug"`
	EndDateISO    *time.Time `json:"end_date_iso,omitempty"`
	GameStartTime *time.Time `json:"game_start_time,omitempty"`
	SecondsDelay  uint64     `json:"seconds_delay"`
	// FPMM is the Fixed Product Market Maker contract address.
	FPMM                 *string         `json:"fpmm,omitempty"`
	MakerBaseFee         decimal.Decimal `json:"maker_base_fee"`
	TakerBaseFee         decimal.Decimal `json:"taker_base_fee"`
	NotificationsEnabled bool            `json:"notifications_enabled"`
	NegRisk              bool            `json:"neg_risk"`
	// NegRiskMarketID is the negative risk market ID (empty string if not a neg risk market).
	NegRiskMarketID *string `json:"neg_risk_market_id,omitempty"`
	// NegRiskRequestID is the negative risk request ID (empty string if not a neg risk market).
	NegRiskRequestID *string  `json:"neg_risk_request_id,omitempty"`
	Icon             string   `json:"icon"`
	Image            string   `json:"image"`
	Rewards          Rewards  `json:"rewards"`
	Is5050Outcome    bool     `json:"is_50_50_outcome"`
	Tokens           []Token  `json:"tokens"`
	Tags             []string `json:"tags"`
}

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
