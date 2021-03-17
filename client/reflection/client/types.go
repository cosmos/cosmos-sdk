package client

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type AccountInfoProvider interface {
	Address(ctx context.Context, pubKey cryptotypes.PubKey) (string, error)
	SigningInfo(ctx context.Context, pubKey cryptotypes.PubKey) (accountNumber, sequence uint64, err error)
	Sign(ctx context.Context, pubKey cryptotypes.PubKey, b []byte) (signedBytes []byte, err error)
}
