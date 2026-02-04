package clob

import "github.com/star-gazer111/poly-go-clob-client/auth"

type L2Client struct {
	*PublicClient
	creds auth.APICreds
}

func NewL2Client(baseURL string, creds auth.APICreds, opts ...PublicClientOption) (*L2Client, error) {
	pc, err := NewPublicClient(baseURL, opts...)
	if err != nil {
		return nil, err
	}

	return &L2Client{
		PublicClient: pc,
		creds:        creds,
	}, nil
}
