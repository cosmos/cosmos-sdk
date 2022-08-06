//go:build rocksdb

package rocksdb

import (
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/db/types"
)

func init() {
	creator := func(name string, dir string) (types.Connection, error) {
		dir = filepath.Join(dir, name)
		return NewDB(dir)
	}
	types.RegisterCreator(types.RocksDBBackend, creator, false)
}
