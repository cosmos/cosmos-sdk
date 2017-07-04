package counter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/txs"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
	eyescli "github.com/tendermint/merkleeyes/client"
	"github.com/tendermint/tmlibs/log"
)

// TODO: actually handle the counter here...
func NewCounterHandler() basecoin.Handler {
	// use the default stack
	coin := coin.NewHandler()
	counter := CounterHandler{}
	dispatcher := stack.NewDispatcher(
		stack.WrapHandler(coin),
		stack.WrapHandler(counter),
	)
	return stack.NewDefault().Use(dispatcher)
}

func TestCounterPlugin(t *testing.T) {
	assert := assert.New(t)

	// Basecoin initialization
	eyesCli := eyescli.NewLocalClient("", 0)
	chainID := "test_chain_id"
	bcApp := app.NewBasecoin(
		NewCounterHandler(),
		eyesCli,
		log.TestingLogger().With("module", "app"),
	)
	bcApp.SetOption("base/chain_id", chainID)
	// t.Log(bcApp.Info())

	// Account initialization
	test1PrivAcc := types.PrivAccountFromSecret("test1")

	// Seed Basecoin with account
	test1Acc := test1PrivAcc.Account
	test1Acc.Balance = types.Coins{{"", 1000}, {"gold", 1000}}
	accOpt, err := json.Marshal(test1Acc)
	require.Nil(t, err)
	log := bcApp.SetOption("coin/account", string(accOpt))
	require.Equal(t, "Success", log)

	// Deliver a CounterTx
	DeliverCounterTx := func(valid bool, counterFee types.Coins, inputSequence int) abci.Result {
		tx := NewCounterTx(valid, counterFee)
		tx = txs.NewChain(chainID, tx)
		stx := txs.NewSig(tx)
		txs.Sign(stx, test1PrivAcc.PrivKey)
		txBytes := wire.BinaryBytes(stx.Wrap())
		return bcApp.DeliverTx(txBytes)
	}

	// Test a basic send, no fee
	res := DeliverCounterTx(true, types.Coins{}, 1)
	assert.True(res.IsOK(), res.String())

	// Test an invalid send, no fee
	res = DeliverCounterTx(false, types.Coins{}, 1)
	assert.True(res.IsErr(), res.String())

	// Test the fee
	res = DeliverCounterTx(true, types.Coins{{"gold", 100}}, 2)
	assert.True(res.IsOK(), res.String())

	// TODO: Test unsupported fee
	// res = DeliverCounterTx(true, types.Coins{{"silver", 100}}, 3)
	// assert.True(res.IsErr(), res.String())
}
