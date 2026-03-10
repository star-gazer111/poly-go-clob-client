package auth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/star-gazer111/poly-go-clob-client/internal/redaction"
)

type Signer interface {
	Address() common.Address
	// signTypedData signs an EIP-712 typed data payload and returns a 65-byte signature
	SignTypedData(ctx context.Context, typedData any) ([]byte, error)
}

type PrivateKeySigner struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

func NewPrivateKeySigner(privateKey *ecdsa.PrivateKey) (*PrivateKeySigner, error) {
	if privateKey == nil {
		return nil, errors.New("private key is nil")
	}

	return &PrivateKeySigner{
		privateKey: privateKey,
		address:    crypto.PubkeyToAddress(privateKey.PublicKey),
	}, nil
}

func NewPrivateKeySignerFromHex(privateKeyHex string) (*PrivateKeySigner, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, err
	}
	return NewPrivateKeySigner(key)
}

func (s *PrivateKeySigner) Address() common.Address {
	return s.address
}

func (s *PrivateKeySigner) SignTypedData(ctx context.Context, typedData any) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if s == nil || s.privateKey == nil {
		return nil, errors.New("private key signer is not initialized")
	}

	td, err := normalizeTypedData(typedData)
	if err != nil {
		return nil, err
	}

	hash, _, err := apitypes.TypedDataAndHash(td)
	if err != nil {
		return nil, err
	}

	return crypto.Sign(hash, s.privateKey)
}

func normalizeTypedData(typedData any) (apitypes.TypedData, error) {
	switch v := typedData.(type) {
	case apitypes.TypedData:
		return v, nil
	case *apitypes.TypedData:
		if v == nil {
			return apitypes.TypedData{}, errors.New("typed data is nil")
		}
		return *v, nil
	default:
		return apitypes.TypedData{}, fmt.Errorf("unsupported typed data type %T", typedData)
	}
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
