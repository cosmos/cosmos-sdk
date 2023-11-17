package account_abstraction

import (
	"context"
	"fmt"

	account_abstractionv1 "cosmossdk.io/api/cosmos/accounts/interfaces/account_abstraction/v1"
	rotationv1 "cosmossdk.io/api/cosmos/accounts/testing/rotation/v1"
	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/api/cosmos/crypto/secp256k1"
	"cosmossdk.io/collections"
	"cosmossdk.io/x/accounts/accountstd"
	"google.golang.org/protobuf/types/known/anypb"

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
		Sequence: collections.NewSequence(d.SchemaBuilder, SequencePrefix, "sequence"),
	}, nil
}

// Full implements the Account interface. It also implements
// the full account abstraction interface.
type Full struct {
	PubKey   collections.Item[*secp256k1.PubKey]
	Sequence collections.Sequence
}

func (a Full) Init(ctx context.Context, msg *rotationv1.MsgInit) (*rotationv1.MsgInitResponse, error) {
	return nil, a.PubKey.Set(ctx, &secp256k1.PubKey{Key: msg.PubKeyBytes})
}

func (a Full) RotatePubKey(ctx context.Context, msg *rotationv1.MsgRotatePubKey) (*rotationv1.MsgRotatePubKeyResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// Authenticate authenticates the account, auth always passess.
func (a Full) Authenticate(_ context.Context, _ *account_abstractionv1.MsgAuthenticate) (*account_abstractionv1.MsgAuthenticateResponse, error) {
	return &account_abstractionv1.MsgAuthenticateResponse{}, nil
}

func (a Full) PayBundler(ctx context.Context, msg *account_abstractionv1.MsgPayBundler) (*account_abstractionv1.MsgPayBundlerResponse, error) {
	// we force this account to pay the bundler only using a bank send.
	if len(msg.BundlerPaymentMessages) != 1 {
		return nil, fmt.Errorf("expected one bundler payment message")
	}
	bankSend, err := accountstd.UnpackAny[bankv1beta1.MsgSend](msg.BundlerPaymentMessages[0])
	if err != nil {
		return nil, err
	}
	if bankSend.FromAddress == "" {
		bankSend.FromAddress = msg.Bundler
	}

	resp, err := accountstd.ExecModule[bankv1beta1.MsgSendResponse](ctx, bankSend)
	if err != nil {
		return nil, err
	}

	anyResp, err := accountstd.PackAny[bankv1beta1.MsgSendResponse](resp)
	if err != nil {
		return nil, err
	}
	return &account_abstractionv1.MsgPayBundlerResponse{
		BundlerPaymentMessagesResponse: []*anypb.Any{anyResp},
	}, nil
}

func (a Full) Execute(ctx context.Context, msg *account_abstractionv1.MsgExecute) (*account_abstractionv1.MsgExecuteResponse, error) {
	// the execute method does ... nothing, just proxies back the requests
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
	accountstd.RegisterExecuteHandler(builder, a.PayBundler)   // implements account_abstraction
	accountstd.RegisterExecuteHandler(builder, a.Execute)      // implements account_abstraction
}

func (a Full) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, a.QueryAuthenticateMethods) // implements account_abstraction
}
