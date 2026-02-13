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
