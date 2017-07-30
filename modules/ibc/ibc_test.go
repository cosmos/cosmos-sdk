package ibc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/light-client/certifiers"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

// this tests registration without registrar permissions
func TestIBCRegister(t *testing.T) {
	assert := assert.New(t)

	// the validators we use to make seeds
	keys := certifiers.GenValKeys(5)
	keys2 := certifiers.GenValKeys(7)
	appHash := []byte{0, 4, 7, 23}
	appHash2 := []byte{12, 34, 56, 78}

	// badSeed doesn't validate
	badSeed := genEmptySeed(keys2, "chain-2", 123, appHash, len(keys2))
	badSeed.Header.AppHash = appHash2

	cases := []struct {
		seed    certifiers.Seed
		checker errors.CheckErr
	}{
		{
			genEmptySeed(keys, "chain-1", 100, appHash, len(keys)),
			errors.NoErr,
		},
		{
			genEmptySeed(keys, "chain-1", 200, appHash, len(keys)),
			IsAlreadyRegisteredErr,
		},
		{
			badSeed,
			IsInvalidCommitErr,
		},
		{
			genEmptySeed(keys2, "chain-2", 123, appHash2, 5),
			errors.NoErr,
		},
	}

	ctx := stack.MockContext("hub", 50)
	store := state.NewMemKVStore()
	app := stack.New().Dispatch(stack.WrapHandler(NewHandler()))

	for i, tc := range cases {
		tx := RegisterChainTx{tc.seed}.Wrap()
		_, err := app.DeliverTx(ctx, store, tx)
		assert.True(tc.checker(err), "%d: %+v", i, err)
	}
}

// this tests permission controls on ibc registration
func TestIBCRegisterPermissions(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// the validators we use to make seeds
	keys := certifiers.GenValKeys(4)
	appHash := []byte{0x17, 0x21, 0x5, 0x1e}

	foobar := basecoin.Actor{App: "foo", Address: []byte("bar")}
	baz := basecoin.Actor{App: "baz", Address: []byte("bar")}
	foobaz := basecoin.Actor{App: "foo", Address: []byte("baz")}

	cases := []struct {
		seed      certifiers.Seed
		registrar basecoin.Actor
		signer    basecoin.Actor
		checker   errors.CheckErr
	}{
		// no sig, no registrar
		{
			seed:    genEmptySeed(keys, "chain-1", 100, appHash, len(keys)),
			checker: errors.NoErr,
		},
		// sig, no registrar
		{
			seed:    genEmptySeed(keys, "chain-2", 100, appHash, len(keys)),
			signer:  foobaz,
			checker: errors.NoErr,
		},
		// registrar, no sig
		{
			seed:      genEmptySeed(keys, "chain-3", 100, appHash, len(keys)),
			registrar: foobar,
			checker:   errors.IsUnauthorizedErr,
		},
		// registrar, wrong sig
		{
			seed:      genEmptySeed(keys, "chain-4", 100, appHash, len(keys)),
			signer:    foobaz,
			registrar: foobar,
			checker:   errors.IsUnauthorizedErr,
		},
		// registrar, wrong sig
		{
			seed:      genEmptySeed(keys, "chain-5", 100, appHash, len(keys)),
			signer:    baz,
			registrar: foobar,
			checker:   errors.IsUnauthorizedErr,
		},
		// registrar, proper sig
		{
			seed:      genEmptySeed(keys, "chain-6", 100, appHash, len(keys)),
			signer:    foobar,
			registrar: foobar,
			checker:   errors.NoErr,
		},
	}

	store := state.NewMemKVStore()
	app := stack.New().Dispatch(stack.WrapHandler(NewHandler()))

	for i, tc := range cases {
		// set option specifies the registrar
		msg, err := json.Marshal(tc.registrar)
		require.Nil(err, "%+v", err)
		_, err = app.InitState(log.NewNopLogger(), store,
			NameIBC, OptionRegistrar, string(msg))
		require.Nil(err, "%+v", err)

		// add permissions to the context
		ctx := stack.MockContext("hub", 50).WithPermissions(tc.signer)
		tx := RegisterChainTx{tc.seed}.Wrap()
		_, err = app.DeliverTx(ctx, store, tx)
		assert.True(tc.checker(err), "%d: %+v", i, err)
	}
}

// this verifies that we can properly update the headers on the chain
func TestIBCUpdate(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// this is the root seed, that others are evaluated against
	keys := certifiers.GenValKeys(7)
	appHash := []byte{0, 4, 7, 23}
	start := 100 // initial height
	root := genEmptySeed(keys, "chain-1", 100, appHash, len(keys))

	keys2 := keys.Extend(2)
	keys3 := keys2.Extend(2)

	// create the app and register the root of trust (for chain-1)
	ctx := stack.MockContext("hub", 50)
	store := state.NewMemKVStore()
	app := stack.New().Dispatch(stack.WrapHandler(NewHandler()))
	tx := RegisterChainTx{root}.Wrap()
	_, err := app.DeliverTx(ctx, store, tx)
	require.Nil(err, "%+v", err)

	cases := []struct {
		seed    certifiers.Seed
		checker errors.CheckErr
	}{
		// same validator, higher up
		{
			genEmptySeed(keys, "chain-1", start+50, []byte{22}, len(keys)),
			errors.NoErr,
		},
		// same validator, between existing (not most recent)
		{
			genEmptySeed(keys, "chain-1", start+5, []byte{15, 43}, len(keys)),
			errors.NoErr,
		},
		// same validators, before root of trust
		{
			genEmptySeed(keys, "chain-1", start-8, []byte{11, 77}, len(keys)),
			IsHeaderNotFoundErr,
		},
		// insufficient signatures
		{
			genEmptySeed(keys, "chain-1", start+60, []byte{24}, len(keys)/2),
			IsInvalidCommitErr,
		},
		// unregistered chain
		{
			genEmptySeed(keys, "chain-2", start+60, []byte{24}, len(keys)/2),
			IsNotRegisteredErr,
		},
		// too much change (keys -> keys3)
		{
			genEmptySeed(keys3, "chain-1", start+100, []byte{22}, len(keys3)),
			IsInvalidCommitErr,
		},
		// legit update to validator set (keys -> keys2)
		{
			genEmptySeed(keys2, "chain-1", start+90, []byte{33}, len(keys2)),
			errors.NoErr,
		},
		// now impossible jump works (keys -> keys2 -> keys3)
		{
			genEmptySeed(keys3, "chain-1", start+100, []byte{44}, len(keys3)),
			errors.NoErr,
		},
	}

	for i, tc := range cases {
		tx := UpdateChainTx{tc.seed}.Wrap()
		_, err := app.DeliverTx(ctx, store, tx)
		assert.True(tc.checker(err), "%d: %+v", i, err)
	}
}

// try to create an ibc packet and verify the number we get back
func TestIBCCreatePacket(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// this is the root seed, that others are evaluated against
	keys := certifiers.GenValKeys(7)
	appHash := []byte{1, 2, 3, 4}
	start := 100 // initial height
	chainID := "cosmos-hub"
	root := genEmptySeed(keys, chainID, start, appHash, len(keys))

	// create the app and register the root of trust (for chain-1)
	ctx := stack.MockContext("hub", 50)
	store := state.NewMemKVStore()
	app := stack.New().Dispatch(stack.WrapHandler(NewHandler()))
	tx := RegisterChainTx{root}.Wrap()
	_, err := app.DeliverTx(ctx, store, tx)
	require.Nil(err, "%+v", err)

	// this is the tx we send, and the needed permission to send it
	raw := stack.NewRawTx([]byte{0xbe, 0xef})
	ibcPerm := AllowIBC(stack.NameOK)
	somePerm := basecoin.Actor{App: "some", Address: []byte("perm")}

	cases := []struct {
		dest     string
		ibcPerms basecoin.Actors
		ctxPerms basecoin.Actors
		checker  errors.CheckErr
	}{
		// wrong chain -> error
		{
			dest:     "some-other-chain",
			ctxPerms: basecoin.Actors{ibcPerm},
			checker:  IsNotRegisteredErr,
		},

		// no ibc permission -> error
		{
			dest:    chainID,
			checker: IsNeedsIBCPermissionErr,
		},

		// correct -> nice sequence
		{
			dest:     chainID,
			ctxPerms: basecoin.Actors{ibcPerm},
			checker:  errors.NoErr,
		},

		// requesting invalid permissions -> error
		{
			dest:     chainID,
			ibcPerms: basecoin.Actors{somePerm},
			ctxPerms: basecoin.Actors{ibcPerm},
			checker:  IsCannotSetPermissionErr,
		},

		// requesting extra permissions when present
		{
			dest:     chainID,
			ibcPerms: basecoin.Actors{somePerm},
			ctxPerms: basecoin.Actors{ibcPerm, somePerm},
			checker:  errors.NoErr,
		},
	}

	for i, tc := range cases {
		tx := CreatePacketTx{
			DestChain:   tc.dest,
			Permissions: tc.ibcPerms,
			Tx:          raw,
		}.Wrap()

		myCtx := ctx.WithPermissions(tc.ctxPerms...)
		_, err = app.DeliverTx(myCtx, store, tx)
		assert.True(tc.checker(err), "%d: %+v", i, err)
	}

	// query packet state - make sure both packets are properly writen
	p := stack.PrefixedStore(NameIBC, store)
	q := OutputQueue(p, chainID)
	if assert.Equal(2, q.Size()) {
		expected := []struct {
			seq  uint64
			perm basecoin.Actors
		}{
			{0, nil},
			{1, basecoin.Actors{somePerm}},
		}

		for _, tc := range expected {
			var packet Packet
			err = wire.ReadBinaryBytes(q.Pop(), &packet)
			require.Nil(err, "%+v", err)
			assert.Equal(chainID, packet.DestChain)
			assert.EqualValues(tc.seq, packet.Sequence)
			assert.Equal(raw, packet.Tx)
			assert.Equal(len(tc.perm), len(packet.Permissions))
		}
	}
}

func TestIBCPostPacket(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	otherID := "chain-1"
	ourID := "hub"
	start := 200
	msg := "it's okay"

	// create the app and our chain
	app := stack.New().
		IBC(NewMiddleware()).
		Dispatch(
			stack.WrapHandler(NewHandler()),
			stack.WrapHandler(stack.OKHandler{Log: msg}),
		)
	ourChain := NewAppChain(app, ourID)

	// set up the other chain and register it with us
	otherChain := NewMockChain(otherID, 7)
	registerTx := otherChain.GetRegistrationTx(start).Wrap()
	_, err := ourChain.DeliverTx(registerTx)
	require.Nil(err, "%+v", err)

	// make a random tx that is to be passed
	rawTx := stack.NewRawTx([]byte{17, 24, 3, 8})

	randomChain := NewMockChain("something-else", 4)
	pbad := NewPacket(rawTx, "something-else", 0)
	packetBad, _ := randomChain.MakePostPacket(pbad, 123)

	p0 := NewPacket(rawTx, ourID, 0)
	packet0, update0 := otherChain.MakePostPacket(p0, start+5)
	require.Nil(ourChain.Update(update0))

	packet0badHeight := packet0
	packet0badHeight.FromChainHeight -= 2

	theirActor := basecoin.Actor{ChainID: otherID, App: "foo", Address: []byte{1}}
	p1 := NewPacket(rawTx, ourID, 1, theirActor)
	packet1, update1 := otherChain.MakePostPacket(p1, start+25)
	require.Nil(ourChain.Update(update1))

	packet1badProof := packet1
	packet1badProof.Key = []byte("random-data")

	ourActor := basecoin.Actor{ChainID: ourID, App: "bar", Address: []byte{2}}
	p2 := NewPacket(rawTx, ourID, 2, ourActor)
	packet2, update2 := otherChain.MakePostPacket(p2, start+50)
	require.Nil(ourChain.Update(update2))

	ibcPerm := basecoin.Actors{AllowIBC(stack.NameOK)}
	cases := []struct {
		packet      PostPacketTx
		permissions basecoin.Actors
		checker     errors.CheckErr
	}{
		// bad chain -> error
		{packetBad, ibcPerm, IsNotRegisteredErr},

		// no matching header -> error
		{packet0badHeight, nil, IsHeaderNotFoundErr},

		// out of order -> error
		{packet1, ibcPerm, IsPacketOutOfOrderErr},

		// all good -> execute tx
		{packet0, ibcPerm, errors.NoErr},

		// bad proof -> error
		{packet1badProof, ibcPerm, IsInvalidProofErr},

		// all good -> execute tx (no special permission needed)
		{packet1, nil, errors.NoErr},

		// repeat -> error
		{packet0, nil, IsPacketAlreadyExistsErr},

		// packet2 contains invalid permissions
		{packet2, nil, IsCannotSetPermissionErr},
	}

	for i, tc := range cases {
		res, err := ourChain.DeliverTx(tc.packet.Wrap(), tc.permissions...)
		assert.True(tc.checker(err), "%d: %+v", i, err)
		if err == nil {
			assert.Equal(msg, res.Log)
		}
	}
}
