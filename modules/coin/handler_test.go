package coin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

// this makes sure that txs are rejected with invalid data or permissions
func TestHandlerValidation(t *testing.T) {
	assert := assert.New(t)

	// these are all valid, except for minusCoins
	addr1 := basecoin.Actor{App: "coin", Address: []byte{1, 2}}
	addr2 := basecoin.Actor{App: "role", Address: []byte{7, 8}}
	someCoins := Coins{{"atom", 123}}
	doubleCoins := Coins{{"atom", 246}}
	minusCoins := Coins{{"eth", -34}}

	cases := []struct {
		valid bool
		tx    basecoin.Tx
		perms []basecoin.Actor
	}{
		// auth works with different apps
		{true,
			NewSendTx(
				[]TxInput{NewTxInput(addr1, someCoins)},
				[]TxOutput{NewTxOutput(addr2, someCoins)}),
			[]basecoin.Actor{addr1}},
		{true,
			NewSendTx(
				[]TxInput{NewTxInput(addr2, someCoins)},
				[]TxOutput{NewTxOutput(addr1, someCoins)}),
			[]basecoin.Actor{addr1, addr2}},
		// check multi-input with both sigs
		{true,
			NewSendTx(
				[]TxInput{NewTxInput(addr1, someCoins), NewTxInput(addr2, someCoins)},
				[]TxOutput{NewTxOutput(addr1, doubleCoins)}),
			[]basecoin.Actor{addr1, addr2}},
		// wrong permissions fail
		{false,
			NewSendTx(
				[]TxInput{NewTxInput(addr1, someCoins)},
				[]TxOutput{NewTxOutput(addr2, someCoins)}),
			[]basecoin.Actor{}},
		{false,
			NewSendTx(
				[]TxInput{NewTxInput(addr1, someCoins)},
				[]TxOutput{NewTxOutput(addr2, someCoins)}),
			[]basecoin.Actor{addr2}},
		{false,
			NewSendTx(
				[]TxInput{NewTxInput(addr1, someCoins), NewTxInput(addr2, someCoins)},
				[]TxOutput{NewTxOutput(addr1, doubleCoins)}),
			[]basecoin.Actor{addr1}},
		// invalid input fails
		{false,
			NewSendTx(
				[]TxInput{NewTxInput(addr1, minusCoins)},
				[]TxOutput{NewTxOutput(addr2, minusCoins)}),
			[]basecoin.Actor{addr2}},
	}

	for i, tc := range cases {
		ctx := stack.MockContext("base-chain", 100).WithPermissions(tc.perms...)
		err := checkTx(ctx, tc.tx.Unwrap().(SendTx))
		if tc.valid {
			assert.Nil(err, "%d: %+v", i, err)
		} else {
			assert.NotNil(err, "%d", i)
		}
	}
}

func TestCheckDeliverSendTx(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// some sample settings
	addr1 := basecoin.Actor{App: "coin", Address: []byte{1, 2}}
	addr2 := basecoin.Actor{App: "role", Address: []byte{7, 8}}
	addr3 := basecoin.Actor{App: "coin", Address: []byte{6, 5, 4, 3}}

	someCoins := Coins{{"atom", 123}}
	moreCoins := Coins{{"atom", 6487}}
	diffCoins := moreCoins.Minus(someCoins)
	otherCoins := Coins{{"eth", 11}}
	mixedCoins := someCoins.Plus(otherCoins)

	type money struct {
		addr  basecoin.Actor
		coins Coins
	}

	cases := []struct {
		init  []money
		tx    basecoin.Tx
		perms []basecoin.Actor
		final []money // nil for error
		cost  uint    // gas allocated (if not error)
	}{
		{
			[]money{{addr1, moreCoins}},
			NewSendTx(
				[]TxInput{NewTxInput(addr1, someCoins)},
				[]TxOutput{NewTxOutput(addr2, someCoins)}),
			[]basecoin.Actor{addr1},
			[]money{{addr1, diffCoins}, {addr2, someCoins}},
			20,
		},
		// simple multi-sig 2 accounts to 1
		{
			[]money{{addr1, mixedCoins}, {addr2, moreCoins}},
			NewSendTx(
				[]TxInput{NewTxInput(addr1, otherCoins), NewTxInput(addr2, someCoins)},
				[]TxOutput{NewTxOutput(addr3, mixedCoins)}),
			[]basecoin.Actor{addr1, addr2},
			[]money{{addr1, someCoins}, {addr2, diffCoins}, {addr3, mixedCoins}},
			30,
		},
		// multi-sig with one account sending many times
		{
			[]money{{addr1, moreCoins.Plus(otherCoins)}},
			NewSendTx(
				[]TxInput{NewTxInput(addr1, otherCoins), NewTxInput(addr1, someCoins)},
				[]TxOutput{NewTxOutput(addr2, mixedCoins)}),
			[]basecoin.Actor{addr1},
			[]money{{addr1, diffCoins}, {addr2, mixedCoins}},
			30,
		},
		// invalid send (not enough money )
		{
			[]money{{addr1, moreCoins}, {addr2, someCoins}},
			NewSendTx(
				[]TxInput{NewTxInput(addr2, moreCoins)},
				[]TxOutput{NewTxOutput(addr1, moreCoins)}),
			[]basecoin.Actor{addr1, addr2},
			nil,
			0,
		},
	}

	h := NewHandler()
	for i, tc := range cases {
		// setup the cases....
		store := state.NewMemKVStore()
		for _, m := range tc.init {
			acct := Account{Coins: m.coins}
			err := storeAccount(store, m.addr.Bytes(), acct)
			require.Nil(err, "%d: %+v", i, err)
		}

		ctx := stack.MockContext("base-chain", 100).WithPermissions(tc.perms...)

		// throw-away state for checktx
		cache := store.Checkpoint()
		cres, err := h.CheckTx(ctx, cache, tc.tx, nil)
		// real store for delivertx
		_, err2 := h.DeliverTx(ctx, store, tc.tx, nil)

		if len(tc.final) > 0 { // valid
			assert.Nil(err, "%d: %+v", i, err)
			assert.Nil(err2, "%d: %+v", i, err2)
			// make sure proper gas is set
			assert.Equal(uint(0), cres.GasPayment, "%d", i)
			assert.Equal(tc.cost, cres.GasAllocated, "%d", i)
			// make sure the final balances are correct
			for _, f := range tc.final {
				acct, err := loadAccount(store, f.addr.Bytes())
				assert.Nil(err, "%d: %+v", i, err)
				assert.Equal(f.coins, acct.Coins)
			}
		} else {
			// both check and deliver should fail
			assert.NotNil(err, "%d", i)
			assert.NotNil(err2, "%d", i)
		}

	}
}

func TestInitState(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// some sample settings
	pk := crypto.GenPrivKeySecp256k1().Wrap()
	addr := pk.PubKey().Address()
	actor := auth.SigPerm(addr)

	someCoins := Coins{{"atom", 123}}
	otherCoins := Coins{{"eth", 11}}
	mixedCoins := someCoins.Plus(otherCoins)

	type money struct {
		addr  basecoin.Actor
		coins Coins
	}

	cases := []struct {
		init     []GenesisAccount
		expected []money
	}{
		{
			[]GenesisAccount{{Address: addr, Balance: mixedCoins}},
			[]money{{actor, mixedCoins}},
		},
	}

	h := NewHandler()
	l := log.NewNopLogger()
	for i, tc := range cases {
		store := state.NewMemKVStore()
		key := "account"

		// set the options
		for j, gen := range tc.init {
			value, err := json.Marshal(gen)
			require.Nil(err, "%d,%d: %+v", i, j, err)
			_, err = h.InitState(l, store, NameCoin, key, string(value), nil)
			require.Nil(err)
		}

		// check state is proper
		for _, f := range tc.expected {
			acct, err := loadAccount(store, f.addr.Bytes())
			assert.Nil(err, "%d: %+v", i, err)
			assert.Equal(f.coins, acct.Coins)
		}
	}
}

func TestSetIssuer(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	cases := []struct {
		issuer basecoin.Actor
	}{
		{basecoin.Actor{App: "sig", Address: []byte("gwkfgk")}},
		// and set back to empty (nil is valid, but assert.Equals doesn't match)
		{basecoin.Actor{Address: []byte{}}},
		{basecoin.Actor{ChainID: "other", App: "role", Address: []byte("vote")}},
	}

	h := NewHandler()
	l := log.NewNopLogger()
	for i, tc := range cases {
		store := state.NewMemKVStore()
		key := "issuer"

		value, err := json.Marshal(tc.issuer)
		require.Nil(err, "%d,%d: %+v", i, err)
		_, err = h.InitState(l, store, NameCoin, key, string(value), nil)
		require.Nil(err, "%+v", err)

		// check state is proper
		info, err := loadHandlerInfo(store)
		assert.Nil(err, "%d: %+v", i, err)
		assert.Equal(tc.issuer, info.Issuer)
	}
}

func TestDeliverCreditTx(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// sample coins
	someCoins := Coins{{"atom", 6570}}
	minusCoins := Coins{{"atom", -1234}}
	lessCoins := someCoins.Plus(minusCoins)
	otherCoins := Coins{{"eth", 11}}
	mixedCoins := someCoins.Plus(otherCoins)

	// some sample addresses
	owner := basecoin.Actor{App: "foo", Address: []byte("rocks")}
	addr1 := basecoin.Actor{App: "coin", Address: []byte{1, 2}}
	key := NewAccountWithKey(someCoins)
	addr2 := key.Actor()
	addr3 := basecoin.Actor{ChainID: "other", App: "sigs", Address: []byte{3, 9}}

	h := NewHandler()
	store := state.NewMemKVStore()
	ctx := stack.MockContext("secret", 77)

	// set the owner who can issue credit
	js, err := json.Marshal(owner)
	require.Nil(err, "%+v", err)
	_, err = h.InitState(log.NewNopLogger(), store, "coin", "issuer", string(js), nil)
	require.Nil(err, "%+v", err)

	// give addr2 some coins to start
	_, err = h.InitState(log.NewNopLogger(), store, "coin", "account", key.MakeOption(), nil)
	require.Nil(err, "%+v", err)

	cases := []struct {
		tx       basecoin.Tx
		perm     basecoin.Actor
		check    errors.CheckErr
		addr     basecoin.Actor
		expected Account
	}{
		// require permission
		{
			tx:    NewCreditTx(addr1, someCoins),
			check: errors.IsUnauthorizedErr,
		},
		// add credit
		{
			tx:       NewCreditTx(addr1, someCoins),
			perm:     owner,
			check:    errors.NoErr,
			addr:     addr1,
			expected: Account{Coins: someCoins, Credit: someCoins},
		},
		// remove some
		{
			tx:       NewCreditTx(addr1, minusCoins),
			perm:     owner,
			check:    errors.NoErr,
			addr:     addr1,
			expected: Account{Coins: lessCoins, Credit: lessCoins},
		},
		// can't remove more cash than there is
		{
			tx:    NewCreditTx(addr1, otherCoins.Negative()),
			perm:  owner,
			check: IsInsufficientFundsErr,
		},
		// cumulative with initial state
		{
			tx:       NewCreditTx(addr2, otherCoins),
			perm:     owner,
			check:    errors.NoErr,
			addr:     addr2,
			expected: Account{Coins: mixedCoins, Credit: otherCoins},
		},
		// Even if there is cash, credit can't go negative
		{
			tx:    NewCreditTx(addr2, minusCoins),
			perm:  owner,
			check: IsInsufficientCreditErr,
		},
		// make sure it works for other chains
		{
			tx:       NewCreditTx(addr3, mixedCoins),
			perm:     owner,
			check:    errors.NoErr,
			addr:     ChainAddr(addr3),
			expected: Account{Coins: mixedCoins, Credit: mixedCoins},
		},
	}

	for i, tc := range cases {
		myStore := store.Checkpoint()

		myCtx := ctx.WithPermissions(tc.perm)
		_, err = h.DeliverTx(myCtx, myStore, tc.tx, nil)
		assert.True(tc.check(err), "%d: %+v", i, err)

		if err == nil {
			store.Commit(myStore)
			acct, err := GetAccount(store, tc.addr)
			require.Nil(err, "%+v", err)
			assert.Equal(tc.expected, acct, "%d", i)
		}
	}
}
