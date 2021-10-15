package orm

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

type TestKeeper struct {
	key             sdk.StoreKey
	autoUInt64Table AutoUInt64Table
	primaryKeyTable PrimaryKeyTable
}

var (
	AutoUInt64TableTablePrefix [2]byte = [2]byte{0x0}
	PrimaryKeyTablePrefix      [2]byte = [2]byte{0x1}
	AutoUInt64TableSeqPrefix   byte    = 0x2
)

func NewTestKeeper(cdc codec.Codec) TestKeeper {
	k := TestKeeper{key: storeKey}

	autoUInt64TableBuilder, err := NewAutoUInt64TableBuilder(AutoUInt64TableTablePrefix, AutoUInt64TableSeqPrefix, &testdata.TableModel{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	k.autoUInt64Table = autoUInt64TableBuilder.Build()

	primaryKeyTableBuilder, err := NewPrimaryKeyTableBuilder(PrimaryKeyTablePrefix, &testdata.TableModel{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	k.primaryKeyTable = primaryKeyTableBuilder.Build()

	return k
}
