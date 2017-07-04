package counter

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/txs"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
	eyescli "github.com/tendermint/merkleeyes/client"
	"github.com/tendermint/tmlibs/log"
)

func TestCounterPlugin(t *testing.T) {
	assert := assert.New(t)

	// Basecoin initialization
	eyesCli := eyescli.NewLocalClient("", 0)
	chainID := "test_chain_id"

	// l := log.TestingLogger().With("module", "app"),
	l := log.NewTMLogger(os.Stdout).With("module", "app")
	// l = log.NewTracingLogger(l)
	bcApp := app.NewBasecoin(
		NewCounterHandler(),
		eyesCli,
		l,
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
		tx := NewCounterTx(valid, counterFee, inputSequence)
		tx = txs.NewChain(chainID, tx)
		stx := txs.NewSig(tx)
		txs.Sign(stx, test1PrivAcc.PrivKey)
		txBytes := wire.BinaryBytes(stx.Wrap())
		return bcApp.DeliverTx(txBytes)
	}

	// Test a basic send, no fee (doesn't update sequence as no money spent)
	res := DeliverCounterTx(true, nil, 1)
	assert.True(res.IsOK(), res.String())

	// Test an invalid send, no fee
	res = DeliverCounterTx(false, nil, 1)
	assert.True(res.IsErr(), res.String())

	// Test the fee (increments sequence)
	res = DeliverCounterTx(true, types.Coins{{"gold", 100}}, 1)
	assert.True(res.IsOK(), res.String())

	// Test unsupported fee
	res = DeliverCounterTx(true, types.Coins{{"silver", 100}}, 2)
	assert.True(res.IsErr(), res.String())
}
