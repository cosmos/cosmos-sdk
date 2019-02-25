package mock

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

const msgRoute = "testMsg"

var (
	numAccts                       = 2
	genCoins                       = sdk.Coins{sdk.NewInt64Coin("foocoin", 77)}
	accs, addrs, pubKeys, privKeys = CreateGenAccounts(numAccts, genCoins)
)

// testMsg is a mock transaction that has a validation which can fail.
type testMsg struct {
	signers     []sdk.AccAddress
	positiveNum int64
}

func (tx testMsg) Route() string                      { return msgRoute }
func (tx testMsg) Type() string                       { return "test" }
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

// getMockApp returns an initialized mock application.
func getMockApp(t *testing.T) *App {
	mApp := NewApp()

	mApp.Router().AddRoute(msgRoute, func(ctx sdk.Context, msg sdk.Msg) (res sdk.Result) { return })
	require.NoError(t, mApp.CompleteSetup())

	return mApp
}

func TestCheckAndDeliverGenTx(t *testing.T) {
	mApp := getMockApp(t)
	mApp.Cdc.RegisterConcrete(testMsg{}, "mock/testMsg", nil)

	SetGenesis(mApp, accs)
	ctxCheck := mApp.BaseApp.NewContext(true, abci.Header{})

	msg := testMsg{signers: []sdk.AccAddress{addrs[0]}, positiveNum: 1}

	acct := mApp.AccountKeeper.GetAccount(ctxCheck, addrs[0])
	require.Equal(t, accs[0], acct.(*auth.BaseAccount))

	SignCheckDeliver(
		t, mApp.Cdc, mApp.BaseApp, []sdk.Msg{msg},
		[]uint64{accs[0].GetAccountNumber()}, []uint64{accs[0].GetSequence()},
		true, true, privKeys[0],
	)

	// Signing a tx with the wrong privKey should result in an auth error
	res := SignCheckDeliver(
		t, mApp.Cdc, mApp.BaseApp, []sdk.Msg{msg},
		[]uint64{accs[1].GetAccountNumber()}, []uint64{accs[1].GetSequence() + 1},
		true, false, privKeys[1],
	)

	require.Equal(t, sdk.CodeUnauthorized, res.Code, res.Log)
	require.Equal(t, sdk.CodespaceRoot, res.Codespace)

	// Resigning the tx with the correct privKey should result in an OK result
	SignCheckDeliver(
		t, mApp.Cdc, mApp.BaseApp, []sdk.Msg{msg},
		[]uint64{accs[0].GetAccountNumber()}, []uint64{accs[0].GetSequence() + 1},
		true, true, privKeys[0],
	)
}

func TestCheckGenTx(t *testing.T) {
	mApp := getMockApp(t)
	mApp.Cdc.RegisterConcrete(testMsg{}, "mock/testMsg", nil)

	SetGenesis(mApp, accs)

	msg1 := testMsg{signers: []sdk.AccAddress{addrs[0]}, positiveNum: 1}
	CheckGenTx(
		t, mApp.BaseApp, []sdk.Msg{msg1},
		[]uint64{accs[0].GetAccountNumber()}, []uint64{accs[0].GetSequence()},
		true, privKeys[0],
	)

	msg2 := testMsg{signers: []sdk.AccAddress{addrs[0]}, positiveNum: -1}
	CheckGenTx(
		t, mApp.BaseApp, []sdk.Msg{msg2},
		[]uint64{accs[0].GetAccountNumber()}, []uint64{accs[0].GetSequence()},
		false, privKeys[0],
	)
}
