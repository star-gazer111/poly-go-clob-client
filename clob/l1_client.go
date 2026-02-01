package clob

import (
	"github.com/star-gazer111/poly-go-clob-client/auth"
)

type L1Client struct {
	*PublicClient
	signer auth.Signer
}

func NewL1Client(baseURL string, signer auth.Signer, opts ...PublicClientOption) *L1Client {
	return &L1Client{
		PublicClient: NewPublicClient(baseURL, opts...),
		signer:       signer,
	}
}
