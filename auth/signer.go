package auth

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type Signer interface {
	Address() common.Address
	// SignTypedData signs an EIP-712 typed-data payload and returns a 65-byte signature.
	SignTypedData(ctx context.Context, typedData any) ([]byte, error)
}

type APICreds struct {
	Key        string
	Secret     string
	Passphrase string
}

func (c APICreds) Redacted() APICreds {
	return APICreds{
		Key:        redact(c.Key),
		Secret:     redact(c.Secret),
		Passphrase: redact(c.Passphrase),
	}
}

func redact(s string) string {
	if len(s) <= 6 {
		return "***"
	}
	return s[:3] + "***" + s[len(s)-3:]
}
