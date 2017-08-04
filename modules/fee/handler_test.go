package fee_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/fee"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

func TestFeeChecks(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	atom := func(i int64) coin.Coin { return coin.Coin{"atom", i} }
	eth := func(i int64) coin.Coin { return coin.Coin{"eth", i} }
	atoms := func(i int64) coin.Coins { return coin.Coins{{"atom", i}} }
	wallet := func(i, j int64) coin.Coins { return coin.Coins{atom(i), eth(j)} }

	// some coin amounts...
	zero := coin.Coin{}
	mixed := wallet(1200, 55)
	pure := atoms(46657)

	// these are some accounts
	collector := basecoin.Actor{App: fee.NameFee, Address: []byte("mine")}
	key1 := coin.NewAccountWithKey(mixed)
	key2 := coin.NewAccountWithKey(pure)
	act1, act2 := key1.Actor(), key2.Actor()

	// set up the apps....
	disp := stack.NewDispatcher(
		// OKHandler will just return success to a RawTx
		stack.WrapHandler(stack.OKHandler{}),
		// coin is needed to handle the IPC call from Fee middleware
		coin.NewHandler(),
	)
	// app1 requires no fees
	app1 := stack.New(fee.NewSimpleFeeMiddleware(atom(0), collector)).Use(disp)
	// app2 requires 2 atom
	app2 := stack.New(fee.NewSimpleFeeMiddleware(atom(2), collector)).Use(disp)

	// set up the store and init the accounts
	store := state.NewMemKVStore()
	l := log.NewNopLogger()
	_, err := app1.InitState(l, store, "coin", "account", key1.MakeOption())
	require.Nil(err, "%+v", err)
	_, err = app2.InitState(l, store, "coin", "account", key2.MakeOption())
	require.Nil(err, "%+v", err)

	// feeCost is what we expect if the fee is properly paid
	feeCost := coin.CostSend * 2
	cases := []struct {
		valid bool
		// this is the middleware stack to test
		app basecoin.Handler
		// they sign the tx
		signer basecoin.Actor
		// wrap with the given fee if hasFee is true
		hasFee bool
		payer  basecoin.Actor
		fee    coin.Coin
		// expected balance after the tx
		left      coin.Coins
		collected coin.Coins
		// expected gas allocated
		expectedCost uint64
	}{
		// make sure it works with no fee (control group)
		{true, app1, act1, false, act1, zero, mixed, nil, 0},
		{true, app1, act2, false, act2, zero, pure, nil, 0},

		// no fee or too low is rejected
		{false, app2, act1, false, act1, zero, mixed, nil, 0},
		{false, app2, act2, false, act2, zero, pure, nil, 0},
		{false, app2, act1, true, act1, zero, mixed, nil, 0},
		{false, app2, act2, true, act2, atom(1), pure, nil, 0},

		// proper fees are transfered in both cases
		{true, app1, act1, true, act1, atom(1), wallet(1199, 55), atoms(1), feeCost},
		{true, app2, act2, true, act2, atom(57), atoms(46600), atoms(58), feeCost},

		// // fee must be the proper type...
		{false, app1, act1, true, act1, eth(5), wallet(1199, 55), atoms(58), 0},

		// signature must match
		{false, app2, act1, true, act2, atom(5), atoms(46600), atoms(58), 0},

		// send only works within limits
		{true, app2, act1, true, act1, atom(1100), wallet(99, 55), atoms(1158), feeCost},
		{false, app2, act1, true, act1, atom(500), wallet(99, 55), atoms(1158), 0},
	}

	for i, tc := range cases {
		// build the tx
		tx := stack.NewRawTx([]byte{7, 8, 9})
		if tc.hasFee {
			tx = fee.NewFee(tx, tc.fee, tc.payer)
		}

		// set up the permissions
		ctx := stack.MockContext("x", 1).WithPermissions(tc.signer)

		// call checktx...
		cres, err := tc.app.CheckTx(ctx, store, tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", i, err)
			assert.EqualValues(tc.fee.Amount, cres.GasPayment)
			assert.EqualValues(tc.expectedCost, cres.GasAllocated)
		} else {
			assert.NotNil(err, "%d", i)
		}

		// call delivertx...
		_, err = tc.app.DeliverTx(ctx, store, tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", i, err)
		} else {
			assert.NotNil(err, "%d", i)
		}

		// check the account balance afterwards....
		cspace := stack.PrefixedStore(coin.NameCoin, store)
		acct, err := coin.GetAccount(cspace, tc.payer)
		require.Nil(err, "%d: %+v", i, err)
		assert.Equal(tc.left, acct.Coins, "%d", i)

		// check the collected balance afterwards....
		acct, err = coin.GetAccount(cspace, collector)
		require.Nil(err, "%d: %+v", i, err)
		assert.Equal(tc.collected, acct.Coins, "%d", i)
	}

}
