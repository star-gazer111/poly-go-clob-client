package types

type OrderBookSummaryRequest struct {
	TokenId uint `json:"token_id"`
	Side    Side `json:"side"`
}

// MarketRequest represents a request to fetch a market by condition ID.
type MarketRequest struct {
	ConditionID string `json:"condition_id"`
}
