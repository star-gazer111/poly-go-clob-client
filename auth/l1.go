package auth

import "context"

// TODO: implement exact Polymarket typed-data and API call to derive creds.
// This should be kept aligned with official clients: "createOrDeriveApiKey".
func DeriveAPICreds(ctx context.Context, signer Signer, baseURL string) (APICreds, error) {
	return APICreds{}, nil
}
