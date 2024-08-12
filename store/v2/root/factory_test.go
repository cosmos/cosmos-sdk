package root

import (
	"testing"

	"github.com/stretchr/testify/require"

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
	require.Error(t, err)
	require.Nil(t, f)
}
