package mock

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

// A mock transaction that has a validation which can fail.
type testMsg struct {
	signers     []sdk.AccAddress
	positiveNum int64
}

// TODO: Clean this up, make it public
const msgType = "testMsg"

func (tx testMsg) Type() string                       { return msgType }
func (tx testMsg) GetMsg() sdk.Msg                    { return tx }
func (tx testMsg) GetMemo() string                    { return "" }
func (tx testMsg) GetSignBytes() []byte               { return nil }
func (tx testMsg) GetSigners() []sdk.AccAddress       { return tx.signers }
func (tx testMsg) GetSignatures() []auth.StdSignature { return nil }
func (tx testMsg) ValidateBasic() sdk.Error {
	if tx.positiveNum >= 0 {
		return nil
	}
	return sdk.ErrTxDecode("positiveNum should be a non-negative integer.")
}

// test auth module messages

var (
	priv1 = crypto.GenPrivKeyEd25519()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	priv2 = crypto.GenPrivKeyEd25519()
	addr2 = sdk.AccAddress(priv2.PubKey().Address())

	coins    = sdk.Coins{sdk.NewCoin("foocoin", 10)}
	testMsg1 = testMsg{signers: []sdk.AccAddress{addr1}, positiveNum: 1}
)

// initialize the mock application for this module
func getMockApp(t *testing.T) *App {
	mapp := NewApp()

	mapp.Router().AddRoute(msgType, func(ctx sdk.Context, msg sdk.Msg) (res sdk.Result) { return })
	require.NoError(t, mapp.CompleteSetup([]*sdk.KVStoreKey{}))
	return mapp
}

func TestMsgPrivKeys(t *testing.T) {
	mapp := getMockApp(t)
	mapp.Cdc.RegisterConcrete(testMsg{}, "mock/testMsg", nil)

	// Construct some genesis bytes to reflect basecoin/types/AppAccount
	// Give 77 foocoin to the first key
	coins := sdk.Coins{sdk.NewCoin("foocoin", 77)}
	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}
	accs := []auth.Account{acc1}

	// Construct genesis state
	SetGenesis(mapp, accs)

	// A checkTx context (true)
	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	res1 := mapp.AccountMapper.GetAccount(ctxCheck, addr1)
	require.Equal(t, acc1, res1.(*auth.BaseAccount))

	// Run a CheckDeliver
	SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{testMsg1}, []int64{0}, []int64{0}, true, priv1)

	// signing a SendMsg with the wrong privKey should be an auth error
	mapp.BeginBlock(abci.RequestBeginBlock{})
	tx := GenTx([]sdk.Msg{testMsg1}, []int64{0}, []int64{1}, priv2)
	res := mapp.Deliver(tx)
	require.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnauthorized), res.Code, res.Log)

	// resigning the tx with the correct priv key should still work
	res = SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{testMsg1}, []int64{0}, []int64{1}, true, priv1)

	require.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeOK), res.Code, res.Log)
}
