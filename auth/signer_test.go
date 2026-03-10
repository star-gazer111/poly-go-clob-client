package auth

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

func TestAPICredsStringIsRedacted(t *testing.T) {
	c := APICreds{
		Key:        "key_abcdefghijklmnopqrstuvwxyz",
		Secret:     "secret_abcdefghijklmnopqrstuvwxyz",
		Passphrase: "pass_abcdefghijklmnopqrstuvwxyz",
	}

	s := c.String()
	// Must not contain full raw values
	if contains(s, c.Key) || contains(s, c.Secret) || contains(s, c.Passphrase) {
		t.Fatalf("String() leaked secret: %s", s)
	}
}

func contains(hay, needle string) bool {
	return needle != "" && len(needle) > 0 && (len(hay) >= len(needle)) && (stringIndex(hay, needle) >= 0)
}

// small helper to avoid pulling strings package into tests? (we can just use strings.Contains)
func stringIndex(s, sub string) int {
	// naive index
	n := len(sub)
	for i := 0; i+n <= len(s); i++ {
		if s[i:i+n] == sub {
			return i
		}
	}
	return -1
}

func TestNewPrivateKeySignerFromHex(t *testing.T) {
	signer, err := NewPrivateKeySignerFromHex("0x4c0883a6910395b37d6231471b5dbb6204fe512961708279f0a4d1d6510c2c9c")
	if err != nil {
		t.Fatalf("NewPrivateKeySignerFromHex() error = %v", err)
	}

	if signer.Address().Hex() == "" {
		t.Fatal("expected signer address to be set")
	}
}

func TestPrivateKeySignerSignTypedData(t *testing.T) {
	signer, err := NewPrivateKeySignerFromHex("0x4c0883a6910395b37d6231471b5dbb6204fe512961708279f0a4d1d6510c2c9c")
	if err != nil {
		t.Fatalf("NewPrivateKeySignerFromHex() error = %v", err)
	}

	td := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
			},
			"Mail": {
				{Name: "contents", Type: "string"},
			},
		},
		PrimaryType: "Mail",
		Domain: apitypes.TypedDataDomain{
			Name: "example",
		},
		Message: apitypes.TypedDataMessage{
			"contents": "hello",
		},
	}

	sig, err := signer.SignTypedData(context.Background(), td)
	if err != nil {
		t.Fatalf("SignTypedData() error = %v", err)
	}
	if len(sig) != 65 {
		t.Fatalf("unexpected signature length: got %d", len(sig))
	}

	hash, _, err := apitypes.TypedDataAndHash(td)
	if err != nil {
		t.Fatalf("TypedDataAndHash() error = %v", err)
	}

	pub, err := crypto.SigToPub(hash, sig)
	if err != nil {
		t.Fatalf("SigToPub() error = %v", err)
	}

	recovered := crypto.PubkeyToAddress(*pub)
	if recovered != signer.Address() {
		t.Fatalf("unexpected recovered address: got %s want %s", recovered.Hex(), signer.Address().Hex())
	}
}
