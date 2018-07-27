package oracle

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/examples/democoin/mock"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

func defaultContext(keys ...sdk.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	for _, key := range keys {
		cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	}
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, nil)
	return ctx
}

type seqOracle struct {
	Seq   int
	Nonce int
}

func (o seqOracle) Type() string {
	return "seq"
}

func (o seqOracle) ValidateBasic() sdk.Error {
	return nil
}

func makeCodec() *wire.Codec {
	var cdc = wire.NewCodec()

	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(Msg{}, "test/Oracle", nil)

	cdc.RegisterInterface((*Payload)(nil), nil)
	cdc.RegisterConcrete(seqOracle{}, "test/oracle/seqOracle", nil)

	cdc.Seal()

	return cdc
}

func seqHandler(ork Keeper, key sdk.StoreKey, codespace sdk.CodespaceType) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case Msg:
			return ork.Handle(func(ctx sdk.Context, p Payload) sdk.Error {
				switch p := p.(type) {
				case seqOracle:
					return handleSeqOracle(ctx, key, p)
				default:
					return sdk.ErrUnknownRequest("")
				}
			}, ctx, msg, codespace)
		default:
			return sdk.ErrUnknownRequest("").Result()
		}
	}
}

func getSequence(ctx sdk.Context, key sdk.StoreKey) int {
	store := ctx.KVStore(key)
	seqbz := store.Get([]byte("seq"))

	var seq int
	if seqbz == nil {
		seq = 0
	} else {
		wire.NewCodec().MustUnmarshalBinary(seqbz, &seq)
	}

	return seq
}

func handleSeqOracle(ctx sdk.Context, key sdk.StoreKey, o seqOracle) sdk.Error {
	store := ctx.KVStore(key)

	seq := getSequence(ctx, key)
	if seq != o.Seq {
		return sdk.NewError(sdk.CodespaceRoot, 1, "")
	}

	bz := wire.NewCodec().MustMarshalBinary(seq + 1)
	store.Set([]byte("seq"), bz)

	return nil
}

func TestOracle(t *testing.T) {
	cdc := makeCodec()

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	addr3 := []byte("addr3")
	addr4 := []byte("addr4")
	valset := &mock.ValidatorSet{[]mock.Validator{
		{addr1, sdk.NewRat(7)},
		{addr2, sdk.NewRat(7)},
		{addr3, sdk.NewRat(1)},
	}}

	key := sdk.NewKVStoreKey("testkey")
	ctx := defaultContext(key)

	bz, err := json.Marshal(valset)
	require.Nil(t, err)
	ctx = ctx.WithBlockHeader(abci.Header{ValidatorsHash: bz})

	ork := NewKeeper(key, cdc, valset, sdk.NewRat(2, 3), 100)
	h := seqHandler(ork, key, sdk.CodespaceRoot)

	// Nonmock.Validator signed, transaction failed
	msg := Msg{seqOracle{0, 0}, []byte("randomguy")}
	res := h(ctx, msg)
	require.False(t, res.IsOK())
	require.Equal(t, 0, getSequence(ctx, key))

	// Less than 2/3 signed, msg not processed
	msg.Signer = addr1
	res = h(ctx, msg)
	require.True(t, res.IsOK())
	require.Equal(t, 0, getSequence(ctx, key))

	// Double signed, transaction failed
	res = h(ctx, msg)
	require.False(t, res.IsOK())
	require.Equal(t, 0, getSequence(ctx, key))

	// More than 2/3 signed, msg processed
	msg.Signer = addr2
	res = h(ctx, msg)
	require.True(t, res.IsOK())
	require.Equal(t, 1, getSequence(ctx, key))

	// Already processed, transaction failed
	msg.Signer = addr3
	res = h(ctx, msg)
	require.False(t, res.IsOK())
	require.Equal(t, 1, getSequence(ctx, key))

	// Less than 2/3 signed, msg not processed
	msg = Msg{seqOracle{100, 1}, addr1}
	res = h(ctx, msg)
	require.True(t, res.IsOK())
	require.Equal(t, 1, getSequence(ctx, key))

	// More than 2/3 signed but payload is invalid
	msg.Signer = addr2
	res = h(ctx, msg)
	require.True(t, res.IsOK())
	require.NotEqual(t, "", res.Log)
	require.Equal(t, 1, getSequence(ctx, key))

	// Already processed, transaction failed
	msg.Signer = addr3
	res = h(ctx, msg)
	require.False(t, res.IsOK())
	require.Equal(t, 1, getSequence(ctx, key))

	// Should handle mock.Validator set change
	valset.AddValidator(mock.Validator{addr4, sdk.NewRat(12)})
	bz, err = json.Marshal(valset)
	require.Nil(t, err)
	ctx = ctx.WithBlockHeader(abci.Header{ValidatorsHash: bz})

	// Less than 2/3 signed, msg not processed
	msg = Msg{seqOracle{1, 2}, addr1}
	res = h(ctx, msg)
	require.True(t, res.IsOK())
	require.Equal(t, 1, getSequence(ctx, key))

	// Less than 2/3 signed, msg not processed
	msg.Signer = addr2
	res = h(ctx, msg)
	require.True(t, res.IsOK())
	require.Equal(t, 1, getSequence(ctx, key))

	// More than 2/3 signed, msg processed
	msg.Signer = addr4
	res = h(ctx, msg)
	require.True(t, res.IsOK())
	require.Equal(t, 2, getSequence(ctx, key))

	// Should handle mock.Validator set change while oracle process is happening
	msg = Msg{seqOracle{2, 3}, addr4}

	// Less than 2/3 signed, msg not processed
	res = h(ctx, msg)
	require.True(t, res.IsOK())
	require.Equal(t, 2, getSequence(ctx, key))

	// Signed mock.Validator is kicked out
	valset.RemoveValidator(addr4)
	bz, err = json.Marshal(valset)
	require.Nil(t, err)
	ctx = ctx.WithBlockHeader(abci.Header{ValidatorsHash: bz})

	// Less than 2/3 signed, msg not processed
	msg.Signer = addr1
	res = h(ctx, msg)
	require.True(t, res.IsOK())
	require.Equal(t, 2, getSequence(ctx, key))

	// More than 2/3 signed, msg processed
	msg.Signer = addr2
	res = h(ctx, msg)
	require.True(t, res.IsOK())
	require.Equal(t, 3, getSequence(ctx, key))
}
