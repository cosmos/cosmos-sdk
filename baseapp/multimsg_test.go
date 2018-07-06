package baseapp

import (
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

// tests multiple msgs of same type from same address in single tx
func TestMultipleBurn(t *testing.T) {
	// Create app.
	app := newTestApp(t.Name())
	capKey := sdk.NewKVStoreKey("key")
	app.MountStoresIAVL(capKey)
	app.SetTxDecoder(func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx auth.StdTx
		fromJSON(txBytes, &tx)
		return tx, nil
	})

	err := app.LoadLatestVersion(capKey)
	if err != nil {
		panic(err)
	}

	app.accountMapper = auth.NewAccountMapper(app.cdc, capKey, &auth.BaseAccount{})

	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, auth.FeeCollectionKeeper{}))

	app.Router().
		AddRoute("burn", newHandleBurn(app.accountMapper)).
		AddRoute("send", newHandleSpend(app.accountMapper))

	app.InitChain(abci.RequestInitChain{})
	app.BeginBlock(abci.RequestBeginBlock{})

	// Set chain-id
	app.deliverState.ctx = app.deliverState.ctx.WithChainID(t.Name())

	priv := makePrivKey("my secret")
	addr := priv.PubKey().Address()

	addCoins(app.accountMapper, app.deliverState.ctx, addr, sdk.Coins{{"foocoin", sdk.NewInt(100)}})

	require.Equal(t, sdk.Coins{{"foocoin", sdk.NewInt(100)}}, app.accountMapper.GetAccount(app.deliverState.ctx, addr).GetCoins(), "Balance did not update")

	msg := testBurnMsg{addr, sdk.Coins{{"foocoin", sdk.NewInt(50)}}}
	tx := GenTx(t.Name(), []sdk.Msg{msg, msg}, []int64{0}, []int64{0}, priv)

	res := app.Deliver(tx)

	require.Equal(t, true, res.IsOK(), res.Log)
	require.Equal(t, sdk.Coins(nil), getCoins(app.accountMapper, app.deliverState.ctx, addr), "Double burn did not work")
}

// tests multiples msgs of same type from different addresses in single tx
func TestBurnMultipleOwners(t *testing.T) {
	// Create app.
	app := newTestApp(t.Name())
	capKey := sdk.NewKVStoreKey("key")
	app.MountStoresIAVL(capKey)
	app.SetTxDecoder(func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx auth.StdTx
		fromJSON(txBytes, &tx)
		return tx, nil
	})

	err := app.LoadLatestVersion(capKey)
	if err != nil {
		panic(err)
	}

	app.accountMapper = auth.NewAccountMapper(app.cdc, capKey, &auth.BaseAccount{})

	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, auth.FeeCollectionKeeper{}))

	app.Router().
		AddRoute("burn", newHandleBurn(app.accountMapper)).
		AddRoute("send", newHandleSpend(app.accountMapper))

	app.InitChain(abci.RequestInitChain{})
	app.BeginBlock(abci.RequestBeginBlock{})

	// Set chain-id
	app.deliverState.ctx = app.deliverState.ctx.WithChainID(t.Name())

	priv1 := makePrivKey("my secret 1")
	addr1 := priv1.PubKey().Address()

	priv2 := makePrivKey("my secret 2")
	addr2 := priv2.PubKey().Address()

	// fund accounts
	addCoins(app.accountMapper, app.deliverState.ctx, addr1, sdk.Coins{{"foocoin", sdk.NewInt(100)}})
	addCoins(app.accountMapper, app.deliverState.ctx, addr2, sdk.Coins{{"foocoin", sdk.NewInt(100)}})

	require.Equal(t, sdk.Coins{{"foocoin", sdk.NewInt(100)}}, getCoins(app.accountMapper, app.deliverState.ctx, addr1), "Balance1 did not update")
	require.Equal(t, sdk.Coins{{"foocoin", sdk.NewInt(100)}}, getCoins(app.accountMapper, app.deliverState.ctx, addr2), "Balance2 did not update")

	msg1 := testBurnMsg{addr1, sdk.Coins{{"foocoin", sdk.NewInt(100)}}}
	msg2 := testBurnMsg{addr2, sdk.Coins{{"foocoin", sdk.NewInt(100)}}}

	// test wrong signers: Address 1 signs both messages
	tx := GenTx(t.Name(), []sdk.Msg{msg1, msg2}, []int64{0, 0}, []int64{0, 0}, priv1, priv1)

	res := app.Deliver(tx)
	require.Equal(t, sdk.ABCICodeType(0x10003), res.Code, "Wrong signatures passed")

	require.Equal(t, sdk.Coins{{"foocoin", sdk.NewInt(100)}}, getCoins(app.accountMapper, app.deliverState.ctx, addr1), "Balance1 changed after invalid sig")
	require.Equal(t, sdk.Coins{{"foocoin", sdk.NewInt(100)}}, getCoins(app.accountMapper, app.deliverState.ctx, addr2), "Balance2 changed after invalid sig")

	// test valid tx
	tx = GenTx(t.Name(), []sdk.Msg{msg1, msg2}, []int64{0, 1}, []int64{1, 0}, priv1, priv2)

	res = app.Deliver(tx)
	require.Equal(t, true, res.IsOK(), res.Log)

	require.Equal(t, sdk.Coins(nil), getCoins(app.accountMapper, app.deliverState.ctx, addr1), "Balance1 did not change after valid tx")
	require.Equal(t, sdk.Coins(nil), getCoins(app.accountMapper, app.deliverState.ctx, addr2), "Balance2 did not change after valid tx")
}

func getCoins(am auth.AccountMapper, ctx sdk.Context, addr sdk.Address) sdk.Coins {
	return am.GetAccount(ctx, addr).GetCoins()
}

func addCoins(am auth.AccountMapper, ctx sdk.Context, addr sdk.Address, coins sdk.Coins) sdk.Error {
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		acc = am.NewAccountWithAddress(ctx, addr)
	}
	err := acc.SetCoins(acc.GetCoins().Plus(coins))
	if err != nil {
		fmt.Println(err)
		return sdk.ErrInternal(err.Error())
	}
	am.SetAccount(ctx, acc)
	return nil
}

// tests different msg types in single tx with different addresses
func TestSendBurn(t *testing.T) {
	// Create app.
	app := newTestApp(t.Name())
	capKey := sdk.NewKVStoreKey("key")
	app.MountStoresIAVL(capKey)
	app.SetTxDecoder(func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx auth.StdTx
		fromJSON(txBytes, &tx)
		return tx, nil
	})

	err := app.LoadLatestVersion(capKey)
	if err != nil {
		panic(err)
	}

	app.accountMapper = auth.NewAccountMapper(app.cdc, capKey, &auth.BaseAccount{})

	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, auth.FeeCollectionKeeper{}))

	app.Router().
		AddRoute("burn", newHandleBurn(app.accountMapper)).
		AddRoute("send", newHandleSpend(app.accountMapper))

	app.InitChain(abci.RequestInitChain{})
	app.BeginBlock(abci.RequestBeginBlock{})

	// Set chain-id
	app.deliverState.ctx = app.deliverState.ctx.WithChainID(t.Name())

	priv1 := makePrivKey("my secret 1")
	addr1 := priv1.PubKey().Address()

	priv2 := makePrivKey("my secret 2")
	addr2 := priv2.PubKey().Address()

	// fund accounts
	addCoins(app.accountMapper, app.deliverState.ctx, addr1, sdk.Coins{{"foocoin", sdk.NewInt(100)}})
	acc := app.accountMapper.NewAccountWithAddress(app.deliverState.ctx, addr2)
	app.accountMapper.SetAccount(app.deliverState.ctx, acc)

	require.Equal(t, sdk.Coins{{"foocoin", sdk.NewInt(100)}}, getCoins(app.accountMapper, app.deliverState.ctx, addr1), "Balance1 did not update")

	sendMsg := testSendMsg{addr1, addr2, sdk.Coins{{"foocoin", sdk.NewInt(50)}}}

	msg1 := testBurnMsg{addr1, sdk.Coins{{"foocoin", sdk.NewInt(50)}}}
	msg2 := testBurnMsg{addr2, sdk.Coins{{"foocoin", sdk.NewInt(50)}}}

	// send then burn
	tx := GenTx(t.Name(), []sdk.Msg{sendMsg, msg2, msg1}, []int64{0, 1}, []int64{0, 0}, priv1, priv2)

	res := app.Deliver(tx)
	require.Equal(t, true, res.IsOK(), res.Log)

	require.Equal(t, sdk.Coins(nil), getCoins(app.accountMapper, app.deliverState.ctx, addr1), "Balance1 did not change after valid tx")
	require.Equal(t, sdk.Coins(nil), getCoins(app.accountMapper, app.deliverState.ctx, addr2), "Balance2 did not change after valid tx")

	// Check that state is only updated if all msgs in tx pass.
	addCoins(app.accountMapper, app.deliverState.ctx, addr1, sdk.Coins{{"foocoin", sdk.NewInt(50)}})

	// burn then send, with fee thats greater than individual tx, but less than combination
	tx = GenTxWithFeeAmt(50000, t.Name(), []sdk.Msg{msg1, sendMsg}, []int64{0}, []int64{1}, priv1)

	res = app.Deliver(tx)
	require.Equal(t, sdk.ABCICodeType(0x1000c), res.Code, "Allowed tx to pass with insufficient funds")

	// Double check that state is correct after Commit.
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	app.BeginBlock(abci.RequestBeginBlock{})
	app.deliverState.ctx = app.deliverState.ctx.WithChainID(t.Name())

	require.Equal(t, sdk.Coins{{"foocoin", sdk.NewInt(50)}}, getCoins(app.accountMapper, app.deliverState.ctx, addr1), "Allowed valid msg to pass in invalid tx")
	require.Equal(t, sdk.Coins(nil), getCoins(app.accountMapper, app.deliverState.ctx, addr2), "Balance2 changed after invalid tx")
}

// Use burn and send msg types to test multiple msgs in one tx
type testBurnMsg struct {
	Addr   sdk.Address
	Amount sdk.Coins
}

const msgType3 = "burn"

func (msg testBurnMsg) Type() string { return msgType3 }
func (msg testBurnMsg) GetSignBytes() []byte {
	bz, _ := json.Marshal(msg)
	return sdk.MustSortJSON(bz)
}
func (msg testBurnMsg) ValidateBasic() sdk.Error {
	if msg.Addr == nil {
		return sdk.ErrInvalidAddress("Cannot use nil as Address")
	}
	return nil
}
func (msg testBurnMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Addr}
}

type testSendMsg struct {
	Sender   sdk.Address
	Receiver sdk.Address
	Amount   sdk.Coins
}

const msgType4 = "send"

func (msg testSendMsg) Type() string { return msgType4 }
func (msg testSendMsg) GetSignBytes() []byte {
	bz, _ := json.Marshal(msg)
	return sdk.MustSortJSON(bz)
}
func (msg testSendMsg) ValidateBasic() sdk.Error {
	if msg.Sender == nil || msg.Receiver == nil {
		return sdk.ErrInvalidAddress("Cannot use nil as Address")
	}
	return nil
}
func (msg testSendMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Sender}
}

// Simple Handlers for burn and send

func newHandleBurn(am auth.AccountMapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx.GasMeter().ConsumeGas(20000, "burning coins")
		burnMsg := msg.(testBurnMsg)
		err := addCoins(am, ctx, burnMsg.Addr, burnMsg.Amount.Negative())
		if err != nil {
			return err.Result()
		}
		return sdk.Result{}
	}
}

func newHandleSpend(am auth.AccountMapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx.GasMeter().ConsumeGas(40000, "spending coins")
		spendMsg := msg.(testSendMsg)
		err := addCoins(am, ctx, spendMsg.Sender, spendMsg.Amount.Negative())
		if err != nil {
			return err.Result()
		}

		err = addCoins(am, ctx, spendMsg.Receiver, spendMsg.Amount)
		if err != nil {
			return err.Result()
		}
		return sdk.Result{}
	}
}

// generate a signed transaction
func GenTx(chainID string, msgs []sdk.Msg, accnums []int64, seq []int64, priv ...crypto.PrivKey) auth.StdTx {
	return GenTxWithFeeAmt(100000, chainID, msgs, accnums, seq, priv...)
}

// generate a signed transaction with the given fee amount
func GenTxWithFeeAmt(feeAmt int64, chainID string, msgs []sdk.Msg, accnums []int64, seq []int64, priv ...crypto.PrivKey) auth.StdTx {
	// make the transaction free
	fee := auth.StdFee{
		sdk.Coins{{"foocoin", sdk.NewInt(0)}},
		feeAmt,
	}

	sigs := make([]auth.StdSignature, len(priv))
	for i, p := range priv {
		sig, err := p.Sign(auth.StdSignBytes(chainID, accnums[i], seq[i], fee, msgs, ""))
		// TODO: replace with proper error handling:
		if err != nil {
			panic(err)
		}
		sigs[i] = auth.StdSignature{
			PubKey:        p.PubKey(),
			Signature:     sig,
			AccountNumber: accnums[i],
			Sequence:      seq[i],
		}
	}
	return auth.NewStdTx(msgs, fee, sigs, "")
}

// spin up simple app for testing
type testApp struct {
	*BaseApp
	accountMapper auth.AccountMapper
}

func newTestApp(name string) testApp {
	return testApp{
		BaseApp: newBaseApp(name),
	}
}

func MakeCodec() *wire.Codec {
	cdc := wire.NewCodec()
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	crypto.RegisterAmino(cdc)
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "cosmos-sdk/BaseAccount", nil)
	cdc.Seal()
	return cdc
}
