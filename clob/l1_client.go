package clob

import (
	"github.com/star-gazer111/poly-go-clob-client/auth"
)

type L1Client struct {
	*PublicClient
	chainID uint64
	signer  auth.Signer
}

func NewL1Client(baseURL string, chainID uint64, signer auth.Signer, opts ...PublicClientOption) *L1Client {
	return &L1Client{
		PublicClient: NewPublicClient(baseURL, opts...),
		chainID:      chainID,
		signer:       signer,
	}
}
