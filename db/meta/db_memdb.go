// +build memdb

package meta

import (
	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/memdb"
)

func memdbConstructor(_, _ string) (dbm.DB, error) {
	return memdb.NewDB(), nil
}

func init() { registerConstructor(MemDBBackend, memdbConstructor, false) }
