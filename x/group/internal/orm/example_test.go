package orm

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

type TestKeeper struct {
	autoUInt64Table *AutoUInt64Table
	primaryKeyTable *PrimaryKeyTable
}

var (
	AutoUInt64TableTablePrefix           [2]byte = [2]byte{0x0}
	PrimaryKeyTablePrefix                [2]byte = [2]byte{0x1}
	AutoUInt64TableSeqPrefix             byte    = 0x2
	AutoUInt64TableModelByNamePrefix     byte    = 0x3
	AutoUInt64TableModelByNumberPrefix   byte    = 0x4
	AutoUInt64TableModelByMetadataPrefix byte    = 0x5
	PrimaryKeyTableModelByNamePrefix     byte    = 0x6
	PrimaryKeyTableModelByNumberPrefix   byte    = 0x7
	PrimaryKeyTableModelByMetadataPrefix byte    = 0x8
)

func NewTestKeeper(cdc codec.Codec) TestKeeper {
	k := TestKeeper{}

	autoUInt64Table, err := NewAutoUInt64Table(AutoUInt64TableTablePrefix, AutoUInt64TableSeqPrefix, &testdata.TableModel{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	k.autoUInt64Table = autoUInt64Table

	primaryKeyTable, err := NewPrimaryKeyTable(PrimaryKeyTablePrefix, &testdata.TableModel{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	k.primaryKeyTable = primaryKeyTable

	return k
}
