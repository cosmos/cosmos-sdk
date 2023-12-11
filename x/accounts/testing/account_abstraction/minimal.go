package account_abstraction

import (
	"context"
	"fmt"

	"cosmossdk.io/api/cosmos/crypto/secp256k1"
	"cosmossdk.io/collections"
	"cosmossdk.io/x/accounts/accountstd"
	account_abstractionv1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	rotationv1 "cosmossdk.io/x/accounts/testing/rotation/v1"

	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	PubKeyPrefix   = collections.NewPrefix(0)
	SequencePrefix = collections.NewPrefix(1)
)

var _ accountstd.Interface = (*MinimalAbstractedAccount)(nil)

func NewMinimalAbstractedAccount(d accountstd.Dependencies) (MinimalAbstractedAccount, error) {
	return MinimalAbstractedAccount{
		PubKey:   collections.NewItem(d.SchemaBuilder, PubKeyPrefix, "pubkey", codec.CollValueV2[secp256k1.PubKey]()),
		Sequence: collections.NewSequence(d.SchemaBuilder, SequencePrefix, "sequence"),
	}, nil
}

// MinimalAbstractedAccount implements the Account interface.
// It implements the minimum required methods.
type MinimalAbstractedAccount struct {
	PubKey   collections.Item[*secp256k1.PubKey]
	Sequence collections.Sequence
}

func (a MinimalAbstractedAccount) Init(ctx context.Context, msg *rotationv1.MsgInit) (*rotationv1.MsgInitResponse, error) {
	return nil, a.PubKey.Set(ctx, &secp256k1.PubKey{Key: msg.PubKeyBytes})
}

func (a MinimalAbstractedAccount) RotatePubKey(ctx context.Context, msg *rotationv1.MsgRotatePubKey) (*rotationv1.MsgRotatePubKeyResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// Authenticate authenticates the account, auth always passess.
func (a MinimalAbstractedAccount) Authenticate(ctx context.Context, msg *account_abstractionv1.MsgAuthenticate) (*account_abstractionv1.MsgAuthenticateResponse, error) {
	if msg.UserOperation.AuthenticationMethod != "secp256k1" {
		return nil, fmt.Errorf("authentication method not supported")
	}
	_, err := a.Sequence.Next(ctx)
	return &account_abstractionv1.MsgAuthenticateResponse{}, err
}

// QueryAuthenticateMethods queries the authentication methods of the account.
func (a MinimalAbstractedAccount) QueryAuthenticateMethods(ctx context.Context, req *account_abstractionv1.QueryAuthenticationMethods) (*account_abstractionv1.QueryAuthenticationMethodsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a MinimalAbstractedAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, a.Init)
}

func (a MinimalAbstractedAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, a.RotatePubKey)
	accountstd.RegisterExecuteHandler(builder, a.Authenticate) // implements account_abstraction
}

func (a MinimalAbstractedAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, a.QueryAuthenticateMethods) // implements account_abstraction
}
