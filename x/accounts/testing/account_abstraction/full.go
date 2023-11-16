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

var _ accountstd.Interface = (*Full)(nil)

func NewFullImpl(d accountstd.Dependencies) (Full, error) {
	return Full{
		PubKey:   collections.NewItem(d.SchemaBuilder, PubKeyPrefix, "pubkey", codec.CollValueV2[secp256k1.PubKey]()),
		Sequence: collections.NewItem(d.SchemaBuilder, SequencePrefix, "sequence", collections.Uint64Value),
	}, nil
}

// Full implements the Account interface. It also implements
// the full account abstraction interface.
type Full struct {
	PubKey   collections.Item[*secp256k1.PubKey]
	Sequence collections.Item[uint64]
}

func (a Full) Init(ctx context.Context, msg *rotationv1.MsgInit) (*rotationv1.MsgInitResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a Full) RotatePubKey(ctx context.Context, msg *rotationv1.MsgRotatePubKey) (*rotationv1.MsgRotatePubKeyResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// Authenticate authenticates the account, auth always passess.
func (a Full) Authenticate(_ context.Context, _ *account_abstractionv1.MsgAuthenticate) (*account_abstractionv1.MsgAuthenticateResponse, error) {
	return &account_abstractionv1.MsgAuthenticateResponse{}, nil
}

func (a Full) PayBundler(ctx context.Context, msg *account_abstractionv1.MsgPayBundler) (*account_abstractionv1.MsgPayBundlerResponse, error) {
	return nil, nil
}

func (a Full) Execute(ctx context.Context, msg *account_abstractionv1.MsgExecute) (*account_abstractionv1.MsgExecuteResponse, error) {
	return nil, nil
}

// QueryAuthenticateMethods queries the authentication methods of the account.
func (a Full) QueryAuthenticateMethods(ctx context.Context, req *account_abstractionv1.QueryAuthenticationMethods) (*account_abstractionv1.QueryAuthenticationMethodsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a Full) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, a.Init)
}

func (a Full) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, a.RotatePubKey)
	accountstd.RegisterExecuteHandler(builder, a.Authenticate) // implements account_abstraction
}

func (a Full) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, a.QueryAuthenticateMethods) // implements account_abstraction
}
