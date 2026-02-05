package clob

import "github.com/star-gazer111/poly-go-clob-client/auth"

type L1Client struct {
	*PublicClient
	signer auth.Signer
}

func NewL1Client(baseURL string, signer auth.Signer, opts ...PublicClientOption) (*L1Client, error) {
	pc, err := NewPublicClient(baseURL, opts...)
	if err != nil {
		return nil, err
	}

	return &L1Client{
		PublicClient: pc,
		signer:       signer,
	}, nil
}
