package mock

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

const msgType = "testMsg"

var (
	numAccts                       = 2
	genCoins                       = sdk.Coins{sdk.NewCoin("foocoin", 77)}
	accs, addrs, pubKeys, privKeys = CreateGenAccounts(numAccts, genCoins)
)

// testMsg is a mock transaction that has a validation which can fail.
type testMsg struct {
	signers     []sdk.AccAddress
	positiveNum int64
}

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

// getMockApp returns an initialized mock application.
func getMockApp(t *testing.T) *App {
	mApp := NewApp()

	mApp.Router().AddRoute(msgType, func(ctx sdk.Context, msg sdk.Msg) (res sdk.Result) { return })
	require.NoError(t, mApp.CompleteSetup([]*sdk.KVStoreKey{}))

	return mApp
}

func TestCheckAndDeliverGenTx(t *testing.T) {
	mApp := getMockApp(t)
	mApp.Cdc.RegisterConcrete(testMsg{}, "mock/testMsg", nil)

	SetGenesis(mApp, accs)
	ctxCheck := mApp.BaseApp.NewContext(true, abci.Header{})

	msg := testMsg{signers: []sdk.AccAddress{addrs[0]}, positiveNum: 1}

	acct := mApp.AccountMapper.GetAccount(ctxCheck, addrs[0])
	require.Equal(t, accs[0], acct.(*auth.BaseAccount))

	SignCheckDeliver(
		t, mApp.BaseApp, []sdk.Msg{msg},
		[]int64{accs[0].GetAccountNumber()}, []int64{accs[0].GetSequence()},
		true, privKeys[0],
	)

	// Signing a tx with the wrong privKey should result in an auth error
	res := SignCheckDeliver(
		t, mApp.BaseApp, []sdk.Msg{msg},
		[]int64{accs[1].GetAccountNumber()}, []int64{accs[1].GetSequence() + 1},
		false, privKeys[1],
	)
	require.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnauthorized), res.Code, res.Log)

	// Resigning the tx with the correct privKey should result in an OK result
	SignCheckDeliver(
		t, mApp.BaseApp, []sdk.Msg{msg},
		[]int64{accs[0].GetAccountNumber()}, []int64{accs[0].GetSequence() + 1},
		true, privKeys[0],
	)
}

func TestCheckGenTx(t *testing.T) {
	mApp := getMockApp(t)
	mApp.Cdc.RegisterConcrete(testMsg{}, "mock/testMsg", nil)

	SetGenesis(mApp, accs)

	msg1 := testMsg{signers: []sdk.AccAddress{addrs[0]}, positiveNum: 1}
	CheckGenTx(
		t, mApp.BaseApp, []sdk.Msg{msg1},
		[]int64{accs[0].GetAccountNumber()}, []int64{accs[0].GetSequence()},
		true, privKeys[0],
	)

	msg2 := testMsg{signers: []sdk.AccAddress{addrs[0]}, positiveNum: -1}
	CheckGenTx(
		t, mApp.BaseApp, []sdk.Msg{msg2},
		[]int64{accs[0].GetAccountNumber()}, []int64{accs[0].GetSequence()},
		false, privKeys[0],
	)
}
