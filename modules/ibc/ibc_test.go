package ibc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/light-client/certifiers"
)

type checkErr func(error) bool

func noErr(err error) bool {
	return err == nil
}

func genEmptySeed(keys certifiers.ValKeys, chain string, h int,
	appHash []byte, count int) certifiers.Seed {

	vals := keys.ToValidators(10, 0)
	cp := keys.GenCheckpoint(chain, h, nil, vals, appHash, 0, count)
	return certifiers.Seed{cp, vals}
}

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
		checker checkErr
	}{
		{
			genEmptySeed(keys, "chain-1", 100, appHash, len(keys)),
			noErr,
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
			noErr,
		},
	}

	ctx := stack.MockContext("hub", 50)
	store := state.NewMemKVStore()
	// no registrar here
	app := stack.New().Dispatch(stack.WrapHandler(NewHandler()))

	for i, tc := range cases {
		tx := RegisterChainTx{tc.seed}.Wrap()
		_, err := app.DeliverTx(ctx, store, tx)
		assert.True(tc.checker(err), "%d: %+v", i, err)
	}
}

func TestIBCUpdate(t *testing.T) {

}

func TestIBCCreatePacket(t *testing.T) {

}

func TestIBCPostPacket(t *testing.T) {

}

func TestIBCSendTx(t *testing.T) {

}
