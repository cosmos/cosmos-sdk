package root

import (
	"testing"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2/db"
)

func TestFactory(t *testing.T) {
	fop := FactoryOptions{
		Logger:    coretesting.NewNopLogger(),
		RootDir:   t.TempDir(),
		Options:   DefaultStoreOptions(),
		StoreKeys: storeKeys,
		SCRawDB:   db.NewMemDB(),
	}

	f, err := CreateRootStore(&fop)
	require.NoError(t, err)
	require.NotNil(t, f)

	fop.Options.SCType = SCTypeIavlV2
	f, err = CreateRootStore(&fop)
	require.NoError(t, err)
	require.NotNil(t, f)

	require.NoError(t, setLatestVersion(fop.SCRawDB, 1))
	fop.Options.SCType = SCTypeIavl
	f, err = CreateRootStore(&fop)
	require.NoError(t, err)
	require.NotNil(t, f)
}

func setLatestVersion(db corestore.KVStoreWithBatch, version int64) error {
	bz, err := gogotypes.StdInt64Marshal(version)
	if err != nil {
		panic(err)
	}
	return db.Set([]byte("s/latest"), bz)
}
