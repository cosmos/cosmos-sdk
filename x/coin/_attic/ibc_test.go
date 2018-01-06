package coin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
	wire "github.com/tendermint/go-wire"
)

// TODO: other test making sure tx is output on send, balance is updated

// This makes sure we respond properly to posttx
// TODO: set credit limit
func TestIBCPostPacket(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	otherID := "chain-2"
	ourID := "dex"
	start := 200

	// create the app and our chain
	app := stack.New().
		IBC(ibc.NewMiddleware()).
		Dispatch(
			NewHandler(),
			stack.WrapHandler(ibc.NewHandler()),
		)
	ourChain := ibc.NewAppChain(app, ourID)

	// set up the other chain and register it with us
	otherChain := ibc.NewMockChain(otherID, 7)
	registerTx := otherChain.GetRegistrationTx(start).Wrap()
	_, err := ourChain.DeliverTx(registerTx)
	require.Nil(err, "%+v", err)

	// set up a rich guy on this chain
	wealth := Coins{{"btc", 300}, {"eth", 2000}, {"ltc", 5000}}
	rich := NewAccountWithKey(wealth)
	_, err = ourChain.InitState("coin", "account", rich.MakeOption())
	require.Nil(err, "%+v", err)

	// sends money to another guy on a different chain, now other chain has credit
	buddy := sdk.Actor{ChainID: otherID, App: auth.NameSigs, Address: []byte("dude")}
	outTx := NewSendOneTx(rich.Actor(), buddy, wealth)
	_, err = ourChain.DeliverTx(outTx, rich.Actor())
	require.Nil(err, "%+v", err)

	// make sure the money moved to the other chain...
	cstore := ourChain.GetStore(NameCoin)
	acct, err := GetAccount(cstore, ChainAddr(buddy))
	require.Nil(err, "%+v", err)
	require.Equal(wealth, acct.Coins)

	// make sure there is a proper packet for this....
	istore := ourChain.GetStore(ibc.NameIBC)
	assertPacket(t, istore, otherID, wealth)

	// these are the people for testing incoming ibc from the other chain
	recipient := sdk.Actor{ChainID: ourID, App: auth.NameSigs, Address: []byte("bar")}
	sender := sdk.Actor{ChainID: otherID, App: auth.NameSigs, Address: []byte("foo")}
	payment := Coins{{"eth", 100}, {"ltc", 300}}
	coinTx := NewSendOneTx(sender, recipient, payment)
	wrongCoin := NewSendOneTx(sender, recipient, Coins{{"missing", 20}})

	p0 := ibc.NewPacket(coinTx, ourID, 0, sender)
	packet0, update0 := otherChain.MakePostPacket(p0, start+5)
	require.Nil(ourChain.Update(update0))

	p1 := ibc.NewPacket(coinTx, ourID, 1, sender)
	packet1, update1 := otherChain.MakePostPacket(p1, start+25)
	require.Nil(ourChain.Update(update1))

	p2 := ibc.NewPacket(wrongCoin, ourID, 2, sender)
	packet2, update2 := otherChain.MakePostPacket(p2, start+50)
	require.Nil(ourChain.Update(update2))

	ibcPerm := sdk.Actors{ibc.AllowIBC(NameCoin)}
	cases := []struct {
		packet      ibc.PostPacketTx
		permissions sdk.Actors
		checker     errors.CheckErr
	}{
		// out of order -> error
		{packet1, ibcPerm, ibc.IsPacketOutOfOrderErr},

		// all good -> execute tx
		{packet0, ibcPerm, errors.NoErr},

		// all good -> execute tx (even if earlier attempt failed)
		{packet1, ibcPerm, errors.NoErr},

		// packet 2 attempts to spend money this chain doesn't have
		{packet2, ibcPerm, IsInsufficientFundsErr},
	}

	for i, tc := range cases {
		_, err := ourChain.DeliverTx(tc.packet.Wrap(), tc.permissions...)
		assert.True(tc.checker(err), "%d: %+v", i, err)
	}

	// now, make sure the recipient got credited for the 2 successful sendtx
	cstore = ourChain.GetStore(NameCoin)
	// FIXME: we need to strip off this when it is local chain-id...
	// think this throw and handle this better
	local := recipient.WithChain("")
	acct, err = GetAccount(cstore, local)
	require.Nil(err, "%+v", err)
	assert.Equal(payment.Plus(payment), acct.Coins)

}

func assertPacket(t *testing.T, istore state.SimpleDB, destID string, amount Coins) {
	assert := assert.New(t)
	require := require.New(t)

	iq := ibc.InputQueue(istore, destID)
	require.Equal(0, iq.Size())

	q := ibc.OutputQueue(istore, destID)
	require.Equal(1, q.Size())
	d := q.Item(0)
	var res ibc.Packet
	err := wire.ReadBinaryBytes(d, &res)
	require.Nil(err, "%+v", err)
	assert.Equal(destID, res.DestChain)
	assert.EqualValues(0, res.Sequence)
	stx, ok := res.Tx.Unwrap().(SendTx)
	if assert.True(ok) {
		assert.Equal(1, len(stx.Outputs))
		assert.Equal(amount, stx.Outputs[0].Coins)
	}
}
