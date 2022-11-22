package iavl

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"strconv"
	"testing"

	db "github.com/cosmos/cosmos-db"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/iavl/mock"
)

func BenchmarkNodeKey(b *testing.B) {
	ndb := &nodeDB{}
	hashes := makeHashes(b, 2432325)
	for i := 0; i < b.N; i++ {
		ndb.nodeKey(hashes[i])
	}
}

func BenchmarkOrphanKey(b *testing.B) {
	ndb := &nodeDB{}
	hashes := makeHashes(b, 2432325)
	for i := 0; i < b.N; i++ {
		ndb.orphanKey(1234, 1239, hashes[i])
	}
}

func BenchmarkTreeString(b *testing.B) {
	tree := makeAndPopulateMutableTree(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink, _ = tree.String()
		require.NotNil(b, sink)
	}

	if sink == nil {
		b.Fatal("Benchmark did not run")
	}
	sink = (interface{})(nil)
}

func TestNewNoDbStorage_StorageVersionInDb_Success(t *testing.T) {
	const expectedVersion = defaultStorageVersionValue

	ctrl := gomock.NewController(t)
	dbMock := mock.NewMockDB(ctrl)

	dbMock.EXPECT().Get(gomock.Any()).Return([]byte(expectedVersion), nil).Times(1)
	dbMock.EXPECT().NewBatch().Return(nil).Times(1)

	ndb := newNodeDB(dbMock, 0, nil)
	require.Equal(t, expectedVersion, ndb.storageVersion)
}

func TestNewNoDbStorage_ErrorInConstructor_DefaultSet(t *testing.T) {
	const expectedVersion = defaultStorageVersionValue

	ctrl := gomock.NewController(t)
	dbMock := mock.NewMockDB(ctrl)

	dbMock.EXPECT().Get(gomock.Any()).Return(nil, errors.New("some db error")).Times(1)
	dbMock.EXPECT().NewBatch().Return(nil).Times(1)

	ndb := newNodeDB(dbMock, 0, nil)
	require.Equal(t, expectedVersion, ndb.getStorageVersion())
}

func TestNewNoDbStorage_DoesNotExist_DefaultSet(t *testing.T) {
	const expectedVersion = defaultStorageVersionValue

	ctrl := gomock.NewController(t)
	dbMock := mock.NewMockDB(ctrl)

	dbMock.EXPECT().Get(gomock.Any()).Return(nil, nil).Times(1)
	dbMock.EXPECT().NewBatch().Return(nil).Times(1)

	ndb := newNodeDB(dbMock, 0, nil)
	require.Equal(t, expectedVersion, ndb.getStorageVersion())
}

func TestSetStorageVersion_Success(t *testing.T) {
	const expectedVersion = fastStorageVersionValue

	db := db.NewMemDB()

	ndb := newNodeDB(db, 0, nil)
	require.Equal(t, defaultStorageVersionValue, ndb.getStorageVersion())

	err := ndb.setFastStorageVersionToBatch()
	require.NoError(t, err)

	latestVersion, err := ndb.getLatestVersion()
	require.NoError(t, err)
	require.Equal(t, expectedVersion+fastStorageVersionDelimiter+strconv.Itoa(int(latestVersion)), ndb.getStorageVersion())
	require.NoError(t, ndb.batch.Write())
}

func TestSetStorageVersion_DBFailure_OldKept(t *testing.T) {
	ctrl := gomock.NewController(t)
	dbMock := mock.NewMockDB(ctrl)
	batchMock := mock.NewMockBatch(ctrl)
	rIterMock := mock.NewMockIterator(ctrl)

	expectedErrorMsg := "some db error"

	expectedFastCacheVersion := 2

	dbMock.EXPECT().Get(gomock.Any()).Return([]byte(defaultStorageVersionValue), nil).Times(1)
	dbMock.EXPECT().NewBatch().Return(batchMock).Times(1)

	// rIterMock is used to get the latest version from disk. We are mocking that rIterMock returns latestTreeVersion from disk
	rIterMock.EXPECT().Valid().Return(true).Times(1)
	rIterMock.EXPECT().Key().Return(rootKeyFormat.Key(expectedFastCacheVersion)).Times(1)
	rIterMock.EXPECT().Close().Return(nil).Times(1)

	dbMock.EXPECT().ReverseIterator(gomock.Any(), gomock.Any()).Return(rIterMock, nil).Times(1)
	batchMock.EXPECT().Set(metadataKeyFormat.Key([]byte(storageVersionKey)), []byte(fastStorageVersionValue+fastStorageVersionDelimiter+strconv.Itoa(expectedFastCacheVersion))).Return(errors.New(expectedErrorMsg)).Times(1)

	ndb := newNodeDB(dbMock, 0, nil)
	require.Equal(t, defaultStorageVersionValue, ndb.getStorageVersion())

	err := ndb.setFastStorageVersionToBatch()
	require.Error(t, err)
	require.Equal(t, expectedErrorMsg, err.Error())
	require.Equal(t, defaultStorageVersionValue, ndb.getStorageVersion())
}

func TestSetStorageVersion_InvalidVersionFailure_OldKept(t *testing.T) {
	ctrl := gomock.NewController(t)
	dbMock := mock.NewMockDB(ctrl)
	batchMock := mock.NewMockBatch(ctrl)

	expectedErrorMsg := errInvalidFastStorageVersion

	invalidStorageVersion := fastStorageVersionValue + fastStorageVersionDelimiter + "1" + fastStorageVersionDelimiter + "2"

	dbMock.EXPECT().Get(gomock.Any()).Return([]byte(invalidStorageVersion), nil).Times(1)
	dbMock.EXPECT().NewBatch().Return(batchMock).Times(1)

	ndb := newNodeDB(dbMock, 0, nil)
	require.Equal(t, invalidStorageVersion, ndb.getStorageVersion())

	err := ndb.setFastStorageVersionToBatch()
	require.Error(t, err)
	require.Equal(t, expectedErrorMsg, err.Error())
	require.Equal(t, invalidStorageVersion, ndb.getStorageVersion())
}

func TestSetStorageVersion_FastVersionFirst_VersionAppended(t *testing.T) {
	db := db.NewMemDB()
	ndb := newNodeDB(db, 0, nil)
	ndb.storageVersion = fastStorageVersionValue
	ndb.latestVersion = 100

	err := ndb.setFastStorageVersionToBatch()
	require.NoError(t, err)
	require.Equal(t, fastStorageVersionValue+fastStorageVersionDelimiter+strconv.Itoa(int(ndb.latestVersion)), ndb.storageVersion)
}

func TestSetStorageVersion_FastVersionSecond_VersionAppended(t *testing.T) {
	db := db.NewMemDB()
	ndb := newNodeDB(db, 0, nil)
	ndb.latestVersion = 100

	storageVersionBytes := []byte(fastStorageVersionValue)
	storageVersionBytes[len(fastStorageVersionValue)-1]++ // increment last byte
	ndb.storageVersion = string(storageVersionBytes)

	err := ndb.setFastStorageVersionToBatch()
	require.NoError(t, err)
	require.Equal(t, string(storageVersionBytes)+fastStorageVersionDelimiter+strconv.Itoa(int(ndb.latestVersion)), ndb.storageVersion)
}

func TestSetStorageVersion_SameVersionTwice(t *testing.T) {
	db := db.NewMemDB()
	ndb := newNodeDB(db, 0, nil)
	ndb.latestVersion = 100

	storageVersionBytes := []byte(fastStorageVersionValue)
	storageVersionBytes[len(fastStorageVersionValue)-1]++ // increment last byte
	ndb.storageVersion = string(storageVersionBytes)

	err := ndb.setFastStorageVersionToBatch()
	require.NoError(t, err)
	newStorageVersion := string(storageVersionBytes) + fastStorageVersionDelimiter + strconv.Itoa(int(ndb.latestVersion))
	require.Equal(t, newStorageVersion, ndb.storageVersion)

	err = ndb.setFastStorageVersionToBatch()
	require.NoError(t, err)
	require.Equal(t, newStorageVersion, ndb.storageVersion)
}

// Test case where version is incorrect and has some extra garbage at the end
func TestShouldForceFastStorageUpdate_DefaultVersion_True(t *testing.T) {
	db := db.NewMemDB()
	ndb := newNodeDB(db, 0, nil)
	ndb.storageVersion = defaultStorageVersionValue
	ndb.latestVersion = 100

	shouldForce, err := ndb.shouldForceFastStorageUpgrade()
	require.False(t, shouldForce)
	require.NoError(t, err)
}

func TestShouldForceFastStorageUpdate_FastVersion_Greater_True(t *testing.T) {
	db := db.NewMemDB()
	ndb := newNodeDB(db, 0, nil)
	ndb.latestVersion = 100
	ndb.storageVersion = fastStorageVersionValue + fastStorageVersionDelimiter + strconv.Itoa(int(ndb.latestVersion+1))

	shouldForce, err := ndb.shouldForceFastStorageUpgrade()
	require.True(t, shouldForce)
	require.NoError(t, err)
}

func TestShouldForceFastStorageUpdate_FastVersion_Smaller_True(t *testing.T) {
	db := db.NewMemDB()
	ndb := newNodeDB(db, 0, nil)
	ndb.latestVersion = 100
	ndb.storageVersion = fastStorageVersionValue + fastStorageVersionDelimiter + strconv.Itoa(int(ndb.latestVersion-1))

	shouldForce, err := ndb.shouldForceFastStorageUpgrade()
	require.True(t, shouldForce)
	require.NoError(t, err)
}

func TestShouldForceFastStorageUpdate_FastVersion_Match_False(t *testing.T) {
	db := db.NewMemDB()
	ndb := newNodeDB(db, 0, nil)
	ndb.latestVersion = 100
	ndb.storageVersion = fastStorageVersionValue + fastStorageVersionDelimiter + strconv.Itoa(int(ndb.latestVersion))

	shouldForce, err := ndb.shouldForceFastStorageUpgrade()
	require.False(t, shouldForce)
	require.NoError(t, err)
}

func TestIsFastStorageEnabled_True(t *testing.T) {
	db := db.NewMemDB()
	ndb := newNodeDB(db, 0, nil)
	ndb.latestVersion = 100
	ndb.storageVersion = fastStorageVersionValue + fastStorageVersionDelimiter + strconv.Itoa(int(ndb.latestVersion))

	require.True(t, ndb.hasUpgradedToFastStorage())
}

func TestIsFastStorageEnabled_False(t *testing.T) {
	db := db.NewMemDB()
	ndb := newNodeDB(db, 0, nil)
	ndb.latestVersion = 100
	ndb.storageVersion = defaultStorageVersionValue

	shouldForce, err := ndb.shouldForceFastStorageUpgrade()
	require.False(t, shouldForce)
	require.NoError(t, err)
}

func makeHashes(b *testing.B, seed int64) [][]byte {
	b.StopTimer()
	rnd := rand.NewSource(seed)
	hashes := make([][]byte, b.N)
	hashBytes := 8 * ((hashSize + 7) / 8)
	for i := 0; i < b.N; i++ {
		hashes[i] = make([]byte, hashBytes)
		for b := 0; b < hashBytes; b += 8 {
			binary.BigEndian.PutUint64(hashes[i][b:b+8], uint64(rnd.Int63()))
		}
		hashes[i] = hashes[i][:hashSize]
	}
	b.StartTimer()
	return hashes
}

func makeAndPopulateMutableTree(tb testing.TB) *MutableTree {
	memDB := db.NewMemDB()
	tree, err := NewMutableTreeWithOpts(memDB, 0, &Options{InitialVersion: 9}, false)
	require.NoError(tb, err)

	for i := 0; i < 1e4; i++ {
		buf := make([]byte, 0, (i/255)+1)
		for j := 0; 1<<j <= i; j++ {
			buf = append(buf, byte((i>>j)&0xff))
		}
		tree.Set(buf, buf)
	}
	_, _, err = tree.SaveVersion()
	require.Nil(tb, err, "Expected .SaveVersion to succeed")
	return tree
}
