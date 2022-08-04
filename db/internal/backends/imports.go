// This is a dummy package used to trigger initialization of backend creators
package backends

import (
	_ "github.com/cosmos/cosmos-sdk/db/badgerdb"
	_ "github.com/cosmos/cosmos-sdk/db/memdb"
)
