package clob

import (
	"github.com/star-gazer111/poly-go-clob-client/auth"
)

type L2Client struct {
	*PublicClient
	creds auth.APICreds
}

func NewL2Client(baseURL string, creds auth.APICreds, opts ...PublicClientOption) *L2Client {
	return &L2Client{
		PublicClient: NewPublicClient(baseURL, opts...),
		creds:        creds,
	}
}
