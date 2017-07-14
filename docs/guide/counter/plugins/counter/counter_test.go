package counter

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/go-wire"
	eyescli "github.com/tendermint/merkleeyes/client"
	"github.com/tendermint/tmlibs/log"
)

func TestCounterPlugin(t *testing.T) {
	assert := assert.New(t)

	// Basecoin initialization
	eyesCli := eyescli.NewLocalClient("", 0)
	chainID := "test_chain_id"

	// logger := log.TestingLogger().With("module", "app"),
	logger := log.NewTMLogger(os.Stdout).With("module", "app")
	// logger = log.NewTracingLogger(logger)
	bcApp := app.NewBasecoin(
		NewHandler("gold"),
		eyesCli,
		logger,
	)
	bcApp.SetOption("base/chain_id", chainID)

	// Account initialization
	bal := coin.Coins{{"", 1000}, {"gold", 1000}}
	acct := coin.NewAccountWithKey(bal)
	log := bcApp.SetOption("coin/account", acct.MakeOption())
	require.Equal(t, "Success", log)

	// Deliver a CounterTx
	DeliverCounterTx := func(valid bool, counterFee coin.Coins, sequence uint32) abci.Result {
		tx := NewTx(valid, counterFee)
		tx = nonce.NewTx(sequence, []basecoin.Actor{acct.Actor()}, tx)
		tx = base.NewChainTx(chainID, 0, tx)
		stx := auth.NewSig(tx)
		auth.Sign(stx, acct.Key)
		txBytes := wire.BinaryBytes(stx.Wrap())
		return bcApp.DeliverTx(txBytes)
	}

	// Test a basic send, no fee
	res := DeliverCounterTx(true, nil, 1)
	assert.True(res.IsOK(), res.String())

	// Test an invalid send, no fee
	res = DeliverCounterTx(false, nil, 2)
	assert.True(res.IsErr(), res.String())

	// Test an invalid send, with supported fee
	res = DeliverCounterTx(true, coin.Coins{{"gold", 100}}, 2)
	assert.True(res.IsOK(), res.String())

	// Test unsupported fee
	res = DeliverCounterTx(true, coin.Coins{{"silver", 100}}, 3)
	assert.True(res.IsErr(), res.String())
}
