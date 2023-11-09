package account_abstraction

import (
	"context"
	"fmt"

	account_abstractionv1 "cosmossdk.io/api/cosmos/accounts/interfaces/account_abstraction/v1"
	rotationv1 "cosmossdk.io/api/cosmos/accounts/testing/rotation/v1"
	"cosmossdk.io/api/cosmos/crypto/secp256k1"
	"cosmossdk.io/collections"
	"cosmossdk.io/x/accounts/accountstd"

	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	PubKeyPrefix   = collections.NewPrefix(0)
	SequencePrefix = collections.NewPrefix(1)
)

var _ accountstd.Interface = (*PartialAccount)(nil)

func NewPartialAccount(d accountstd.Dependencies) (PartialAccount, error) {
	return PartialAccount{
		PubKey:   collections.NewItem(d.SchemaBuilder, PubKeyPrefix, "pubkey", codec.CollValueV2[secp256k1.PubKey]()),
		Sequence: collections.NewItem(d.SchemaBuilder, SequencePrefix, "sequence", collections.Uint64Value),
	}, nil
}

// PartialAccount implements the Account interface. It also
// implements the account_abstraction interface, it only implements
// the minimum methods required to be a valid account_abstraction
// implementer.
type PartialAccount struct {
	PubKey   collections.Item[*secp256k1.PubKey]
	Sequence collections.Item[uint64]
}

func (a PartialAccount) Init(ctx context.Context, msg *rotationv1.MsgInit) (*rotationv1.MsgInitResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a PartialAccount) RotatePubKey(ctx context.Context, msg *rotationv1.MsgRotatePubKey) (*rotationv1.MsgRotatePubKeyResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// Authenticate authenticates the account.
func (a PartialAccount) Authenticate(ctx context.Context, msg *account_abstractionv1.MsgAuthenticate) (*account_abstractionv1.MsgAuthenticateResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// QueryAuthenticateMethods queries the authentication methods of the account.
func (a PartialAccount) QueryAuthenticateMethods(ctx context.Context, req *account_abstractionv1.QueryAuthenticationMethods) (*account_abstractionv1.QueryAuthenticationMethodsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a PartialAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, a.Init)
}

func (a PartialAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, a.RotatePubKey)
	accountstd.RegisterExecuteHandler(builder, a.Authenticate) // implements account_abstraction
}

func (a PartialAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, a.QueryAuthenticateMethods) // implements account_abstraction
}
