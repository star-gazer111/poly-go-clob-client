package types

type OrderBookSummaryRequest struct {
	TokenId string `json:"token_id"`
	Side    Side   `json:"side,omitempty"`
}

// MarketRequest represents a request to fetch a market by condition ID.
type MarketRequest struct {
	ConditionID string `json:"condition_id"`
}

type LastTradePriceRequest struct {
	TokenId string `json:"token_id"`
}

type GetMarketTradesEventsRequest struct {
	ConditionID string `json:"condition_id"`
	NextCursor  string `json:"next_cursor,omitempty"`
}

type MidpointRequest struct {
	TokenId string `json:"token_id"`
}

type PriceRequest struct {
	TokenId string `json:"token_id"`
	Side    string `json:"side"` // "BUY" or "SELL"
}
type SpreadRequest struct {
	TokenId string `json:"token_id"`
	Side    *Side  `json:"side,omitempty"`
}

type PricesHistoryRequest struct {
	Market   string   `json:"market"`
	Interval Interval `json:"interval,omitempty"`
	StartTs  *int64   `json:"startTs,omitempty"`
	EndTs    *int64   `json:"endTs,omitempty"`
	Fidelity *uint32  `json:"fidelity,omitempty"`
}
