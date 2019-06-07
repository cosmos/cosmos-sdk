package mapping

import (
	//	"testing"
	"math/rand"

	//	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const testsize = 10

type test interface {
	test()
}

type teststruct struct {
	I  uint64
	B  bool
	SL []byte
}

var _ test = teststruct{}

func (teststruct) test() {}

func newtest() test {
	var res teststruct
	res.I = rand.Uint64()
	res.B = rand.Int()%2 == 0
	res.SL = make([]byte, 20)
	rand.Read(res.SL)
	return res
}

func defaultComponents() (sdk.StoreKey, Context, *codec.Codec) {
	key := sdk.NewKVStoreKey("test")
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	cdc := codec.New()
	cdc.RegisterInterface((*test)(nil), nil)
	cdc.RegisterConcrete(teststruct{}, "test/struct", nil)
	cdc.Seal()
	return key, ctx, cdc
}
