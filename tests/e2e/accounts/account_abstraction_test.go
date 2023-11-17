package accounts

import (
	"testing"

	rotationv1 "cosmossdk.io/api/cosmos/accounts/testing/rotation/v1"
	accountsv1 "cosmossdk.io/api/cosmos/accounts/v1"
	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	privKey    = secp256k1.GenPrivKey()
	accCreator = []byte("creator")
	bundler    = "bundler"
	alice      = "alice"
)

func TestAccountAbstraction(t *testing.T) {
	app := simapp.NewSimApp(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
	ak := app.AccountsKeeper
	ctx := sdk.NewContext(app.CommitMultiStore(), false, app.Logger())

	_, aaAddr, err := ak.Init(ctx, "aa_full", accCreator, &rotationv1.MsgInit{
		PubKeyBytes: privKey.PubKey().Bytes(),
	})
	require.NoError(t, err)

	aaAddrStr, err := app.AuthKeeper.AddressCodec().BytesToString(aaAddr)
	require.NoError(t, err)

	t.Run("ok", func(t *testing.T) {
		resp := ak.ExecuteUserOperation(ctx, bundler, &accountsv1.UserOperation{
			Sender:                 aaAddrStr,
			AuthenticationMethod:   "standard",
			AuthenticationData:     []byte("signature"),
			AuthenticationGasLimit: 10000,
			BundlerPaymentMessages: nil,
			BundlerPaymentGasLimit: 10000,
			ExecutionMessages: intoAny(t, &bankv1beta1.MsgSend{
				FromAddress: "",
				ToAddress:   "",
				Amount:      nil,
			}),
			ExecutionGasLimit: 10000,
		})
		t.Log(resp.String())
	})
	t.Run("ok pay bundler not implemented", func(t *testing.T) {})
	t.Run("ok exec messages not implemented", func(t *testing.T) {})
	t.Run("pay bundle impersonation", func(t *testing.T) {})
	t.Run("exec message impersonation", func(t *testing.T) {})
	t.Run("auth failure", func(t *testing.T) {})
	t.Run("pay bundle failure", func(t *testing.T) {})
	t.Run("exec message failure", func(t *testing.T) {})
}

func intoAny(t *testing.T, msgs ...proto.Message) (anys []*anypb.Any) {
	t.Helper()
	for _, msg := range msgs {
		any, err := anypb.New(msg)
		require.NoError(t, err)
		anys = append(anys, any)
	}
	return
}
