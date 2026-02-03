package auth

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/star-gazer111/poly-go-clob-client/internal/redaction"
)

type Signer interface {
	Address() common.Address
	// signTypedData signs an EIP-712 typed data payload and returns a 65-byte signature
	SignTypedData(ctx context.Context, typedData any) ([]byte, error)
}

type APICreds struct {
	Key        string
	Secret     string
	Passphrase string
}

func (c APICreds) Redacted() APICreds {
	return APICreds{
		Key:        redaction.Redact(c.Key),
		Secret:     redaction.Redact(c.Secret),
		Passphrase: redaction.Redact(c.Passphrase),
	}
}

// String implements fmt.Stringer for safe logging & never returns raw secrets
func (c APICreds) String() string {
	r := c.Redacted()
	return fmt.Sprintf("APICreds{Key=%q Secret=%q Passphrase=%q}", r.Key, r.Secret, r.Passphrase)
}
