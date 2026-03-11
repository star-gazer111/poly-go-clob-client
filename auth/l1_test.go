package auth

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestNewL1AuthPayload(t *testing.T) {
	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	payload := NewL1AuthPayload(address, 1700000000, 7)

	if payload.Address != address {
		t.Fatalf("unexpected address: got %s want %s", payload.Address.Hex(), address.Hex())
	}
	if payload.Timestamp != "1700000000" {
		t.Fatalf("unexpected timestamp: got %s", payload.Timestamp)
	}
	if payload.Nonce != 7 {
		t.Fatalf("unexpected nonce: got %d", payload.Nonce)
	}
	if payload.Message != ClobAuthMessage {
		t.Fatalf("unexpected message: got %q", payload.Message)
	}
}

func TestBuildClobAuthTypedData(t *testing.T) {
	address := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	payload := NewL1AuthPayload(address, 1700000000, 7)

	td, err := BuildClobAuthTypedData(137, payload)
	if err != nil {
		t.Fatalf("BuildClobAuthTypedData() error = %v", err)
	}

	if td.PrimaryType != ClobAuthType {
		t.Fatalf("unexpected primary type: got %s", td.PrimaryType)
	}
	if td.Domain.Name != ClobAuthDomainName {
		t.Fatalf("unexpected domain name: got %s", td.Domain.Name)
	}
	if td.Domain.Version != ClobAuthVersion {
		t.Fatalf("unexpected domain version: got %s", td.Domain.Version)
	}
	if got := td.Message["address"]; got != address.Hex() {
		t.Fatalf("unexpected address field: got %v want %s", got, address.Hex())
	}
	if got := td.Message["timestamp"]; got != "1700000000" {
		t.Fatalf("unexpected timestamp field: got %v", got)
	}
	if got := td.Message["nonce"]; got != "7" {
		t.Fatalf("unexpected nonce field: got %v", got)
	}
	if got := td.Message["message"]; got != ClobAuthMessage {
		t.Fatalf("unexpected message field: got %v", got)
	}
	if len(td.Types[ClobAuthType]) != 4 {
		t.Fatalf("unexpected auth type field count: got %d", len(td.Types[ClobAuthType]))
	}
}
