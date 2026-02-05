package types

type OrderBookSummaryRequest struct {
	TokenId uint `json:"token_id"`
	Side    Side `json:"side"`
}
