package auth

import (
	"context"
	"errors"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	gmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const (
	ClobAuthDomainName = "ClobAuthDomain"
	ClobAuthVersion    = "1"
	ClobAuthType       = "ClobAuth"
	ClobAuthMessage    = "This message attests that I control the given wallet"
)

type L1AuthPayload struct {
	Address   common.Address
	Timestamp string
	Nonce     uint64
	Message   string
}

func NewL1AuthPayload(address common.Address, timestamp int64, nonce uint64) L1AuthPayload {
	return L1AuthPayload{
		Address:   address,
		Timestamp: strconv.FormatInt(timestamp, 10),
		Nonce:     nonce,
		Message:   ClobAuthMessage,
	}
}

func BuildClobAuthTypedData(chainID uint64, payload L1AuthPayload) (apitypes.TypedData, error) {
	if chainID == 0 {
		return apitypes.TypedData{}, errors.New("chain id must be greater than zero")
	}
	if payload.Address == (common.Address{}) {
		return apitypes.TypedData{}, errors.New("address is required")
	}
	if payload.Timestamp == "" {
		return apitypes.TypedData{}, errors.New("timestamp is required")
	}
	if payload.Message == "" {
		return apitypes.TypedData{}, errors.New("message is required")
	}

	chainIDValue := gmath.NewHexOrDecimal256(0)
	(*big.Int)(chainIDValue).SetUint64(chainID)

	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			ClobAuthType: {
				{Name: "address", Type: "address"},
				{Name: "timestamp", Type: "string"},
				{Name: "nonce", Type: "uint256"},
				{Name: "message", Type: "string"},
			},
		},
		PrimaryType: ClobAuthType,
		Domain: apitypes.TypedDataDomain{
			Name:    ClobAuthDomainName,
			Version: ClobAuthVersion,
			ChainId: chainIDValue,
		},
		Message: apitypes.TypedDataMessage{
			"address":   payload.Address.Hex(),
			"timestamp": payload.Timestamp,
			"nonce":     strconv.FormatUint(payload.Nonce, 10),
			"message":   payload.Message,
		},
	}, nil
}

// TODO: implement exact Polymarket typed-data and API call to derive creds.
// This should be kept aligned with official clients: "createOrDeriveApiKey".
func DeriveAPICreds(ctx context.Context, signer Signer, baseURL string) (APICreds, error) {
	return APICreds{}, nil
}
