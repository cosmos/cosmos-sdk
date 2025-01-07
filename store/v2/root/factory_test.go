package root

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/internal/encoding"
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
	require.Nil(t, f)

	require.NoError(t, setLatestVersion(fop.SCRawDB, 1))
	fop.Options.SCType = SCTypeIavl
	f, err = CreateRootStore(&fop)
	require.NoError(t, err)
	require.NotNil(t, f)
	require.True(t, f.(*Store).isMigrating)
}

func setLatestVersion(db corestore.KVStoreWithBatch, version uint64) error {
	var buf bytes.Buffer
	buf.Grow(encoding.EncodeUvarintSize(version))
	if err := encoding.EncodeUvarint(&buf, version); err != nil {
		return err
	}
	return db.Set([]byte("s/latest"), buf.Bytes())
}
