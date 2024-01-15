package base

import (
	"bytes"
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/accounts/accountstd"
	v1 "cosmossdk.io/x/accounts/defaults/base/v1"
	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

var (
	PubKeyPrefix   = collections.NewPrefix(0)
	SequencePrefix = collections.NewPrefix(1)
)

func NewAccount(deps accountstd.Dependencies) (Account, error) {
	return Account{
		PubKey:   collections.NewItem(deps.SchemaBuilder, PubKeyPrefix, "pub_key", codec.CollValue[secp256k1.PubKey](nil)),
		Sequence: collections.NewSequence(deps.SchemaBuilder, SequencePrefix, "sequence"),

		addrCodec: deps.AddressCodec,
	}, nil
}

// Account is the implementation of the modernized auth.BaseAccount.
type Account struct {
	PubKey   collections.Item[secp256k1.PubKey]
	Sequence collections.Sequence

	hs header.Service

	addrCodec        address.Codec
	signModeHandlers *txsigning.HandlerMap
}

func (a Account) Init(ctx context.Context, msg *v1.MsgInit) (*v1.MsgInitResponse, error) {
	return &v1.MsgInitResponse{}, a.validateAndSetPubKey(ctx, msg.PubKey)
}

func (a Account) SwapPubKey(ctx context.Context, msg *v1.MsgSwapPubKey) (*v1.MsgSwapPubKeyResponse, error) {
	if !bytes.Equal(accountstd.Sender(ctx), accountstd.Whoami(ctx)) {
		return nil, fmt.Errorf("unauthorized")
	}
	return &v1.MsgSwapPubKeyResponse{}, a.validateAndSetPubKey(ctx, msg.NewPubKey)
}

func (a Account) validateAndSetPubKey(ctx context.Context, key []byte) error {
	// TODO: add secp256k1 pub key validation.
	return a.PubKey.Set(ctx, secp256k1.PubKey{Key: key})
}

func (a Account) RegisterInitHandler(r *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(r, a.Init)
}

func (a Account) RegisterExecuteHandlers(r *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(r, a.SwapPubKey)
	accountstd.RegisterExecuteHandler(r, a.Authenticate)
}

func (a Account) RegisterQueryHandlers(r *accountstd.QueryBuilder) {}
