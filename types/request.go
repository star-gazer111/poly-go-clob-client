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
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
	TakerOnly   *bool  `json:"taker_only,omitempty"`
	Side        string `json:"side,omitempty"`
}
