package commitment

import (
	"testing"

	dbm "cosmossdk.io/store/v2/db"
	"github.com/stretchr/testify/assert"
)

func TestMetadataStore_GetLatestVersion(t *testing.T) {
	db := dbm.NewMemDB()
	ms := NewMetadataStore(db)

	version, err := ms.GetLatestVersion()
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), version)

	// set latest version
	err = ms.setLatestVersion(10)
	assert.NoError(t, err)

	version, err = ms.GetLatestVersion()
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), version)
}

func TestMetadataStore_GetV2MigrationHeight(t *testing.T) {
	db := dbm.NewMemDB()
	ms := NewMetadataStore(db)

	version, err := ms.GetV2MigrationHeight()
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), version)

	err = ms.setV2MigrationHeight(10)
	assert.NoError(t, err)

	version, err = ms.GetV2MigrationHeight()
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), version)
}
