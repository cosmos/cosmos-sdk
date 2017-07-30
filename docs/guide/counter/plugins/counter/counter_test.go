package counter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/nonce"
)

func TestCounterPlugin(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Basecoin initialization
	chainID := "test_chain_id"
	logger := log.TestingLogger()
	// logger := log.NewTracingLogger(log.NewTMLogger(os.Stdout))

	store, err := app.NewStore("", 0, logger.With("module", "store"))
	require.Nil(err, "%+v", err)

	h := NewHandler("gold")
	bcApp := app.NewBasecoin(
		h,
		store,
		logger.With("module", "app"),
	)
	bcApp.InitState("base/chain_id", chainID)

	// Account initialization
	bal := coin.Coins{{"", 1000}, {"gold", 1000}}
	acct := coin.NewAccountWithKey(bal)
	log := bcApp.InitState("coin/account", acct.MakeOption())
	require.Equal("Success", log)

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

	// Test an invalid sequence
	res = DeliverCounterTx(true, nil, 2)
	assert.True(res.IsErr(), res.String())

	// Test an valid send, with supported fee
	res = DeliverCounterTx(true, coin.Coins{{"gold", 100}}, 3)
	assert.True(res.IsOK(), res.String())

	// Test unsupported fee
	res = DeliverCounterTx(true, coin.Coins{{"silver", 100}}, 4)
	assert.True(res.IsErr(), res.String())
}
