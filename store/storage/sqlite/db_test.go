package sqlite

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/storage"
)

const (
	storeKey1 = "store1"
)

func TestStorageTestSuite(t *testing.T) {
	s := &storage.StorageTestSuite{
		NewDB: func(dir string) (store.VersionedDatabase, error) {
			return New(dir)
		},
		EmptyBatchSize: 0,
	}
	suite.Run(t, s)
}

func TestDatabase_ReverseIterator(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	cs := new(store.Changeset)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		cs.AddKVPair(store.KVPair{StoreKey: storeKey1, Key: []byte(key), Value: []byte(val)})
	}

	require.NoError(t, db.ApplyChangeset(1, cs))

	// reverse iterator without an end key
	iter, err := db.ReverseIterator(storeKey1, 1, []byte("key000"), nil)
	require.NoError(t, err)

	defer iter.Close()

	i, count := 99, 0
	for ; iter.Valid(); iter.Next() {
		require.Equal(t, []byte(fmt.Sprintf("key%03d", i)), iter.Key())
		require.Equal(t, []byte(fmt.Sprintf("val%03d", i)), iter.Value())

		i--
		count++
	}
	require.Equal(t, 100, count)
	require.NoError(t, iter.Error())

	// seek past domain, which should make the iterator invalid and produce an error
	require.False(t, iter.Next())
	require.False(t, iter.Valid())

	// reverse iterator with with a start and end domain
	iter2, err := db.ReverseIterator(storeKey1, 1, []byte("key010"), []byte("key019"))
	require.NoError(t, err)

	defer iter2.Close()

	i, count = 18, 0
	for ; iter2.Valid(); iter2.Next() {
		require.Equal(t, []byte(fmt.Sprintf("key%03d", i)), iter2.Key())
		require.Equal(t, []byte(fmt.Sprintf("val%03d", i)), iter2.Value())

		i--
		count++
	}
	require.Equal(t, 9, count)
	require.NoError(t, iter2.Error())

	// seek past domain, which should make the iterator invalid and produce an error
	require.False(t, iter2.Next())
	require.False(t, iter2.Valid())

	// start must be <= end
	iter3, err := db.ReverseIterator(storeKey1, 1, []byte("key020"), []byte("key019"))
	require.Error(t, err)
	require.Nil(t, iter3)
}

func TestParallelWrites(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	latestVersion := 10
	kvCount := 100

	wg := sync.WaitGroup{}
	triggerStartCh := make(chan bool)

	// start 10 goroutines that write to the database
	for i := 0; i < latestVersion; i++ {
		wg.Add(1)
		go func(i int) {
			<-triggerStartCh
			t.Log("start time", i, time.Now())
			defer wg.Done()
			cs := new(store.Changeset)
			for j := 0; j < kvCount; j++ {
				key := fmt.Sprintf("key-%d-%03d", i, j)
				val := fmt.Sprintf("val-%d-%03d", i, j)

				cs.AddKVPair(store.KVPair{StoreKey: storeKey1, Key: []byte(key), Value: []byte(val)})
			}

			require.NoError(t, db.ApplyChangeset(uint64(i+1), cs))
			t.Log("end time", i, time.Now())
		}(i)

	}

	// start the goroutines
	close(triggerStartCh)
	wg.Wait()

	// check that all the data is there
	for i := 0; i < latestVersion; i++ {
		for j := 0; j < kvCount; j++ {
			version := uint64(i + 1)
			key := fmt.Sprintf("key-%d-%03d", i, j)
			val := fmt.Sprintf("val-%d-%03d", i, j)

			v, err := db.Get(storeKey1, version, []byte(key))
			require.NoError(t, err)
			require.Equal(t, []byte(val), v)
		}
	}
}

func TestParallelWriteAndPruning(t *testing.T) {
	db, err := New(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	latestVersion := 100
	kvCount := 100
	prunePeriod := 5

	wg := sync.WaitGroup{}
	triggerStartCh := make(chan bool)

	// start a goroutine that write to the database
	wg.Add(1)
	go func() {
		<-triggerStartCh
		defer wg.Done()
		for i := 0; i < latestVersion; i++ {
			cs := new(store.Changeset)
			for j := 0; j < kvCount; j++ {
				key := fmt.Sprintf("key-%d-%03d", i, j)
				val := fmt.Sprintf("val-%d-%03d", i, j)

				cs.AddKVPair(store.KVPair{StoreKey: storeKey1, Key: []byte(key), Value: []byte(val)})
			}

			require.NoError(t, db.ApplyChangeset(uint64(i+1), cs))
		}
	}()
	// start a goroutine that prunes the database
	wg.Add(1)
	go func() {
		<-triggerStartCh
		defer wg.Done()
		for i := 10; i < latestVersion; i += prunePeriod {
			for {
				v, err := db.GetLatestVersion()
				require.NoError(t, err)
				if v > uint64(i) {
					t.Log("pruning version", v-1)
					require.NoError(t, db.Prune(v-1))
					break
				}
			}
		}
	}()

	// start the goroutines
	close(triggerStartCh)
	wg.Wait()

	// check if the data is pruned
	version := uint64(latestVersion - prunePeriod)
	val, err := db.Get(storeKey1, version, []byte(fmt.Sprintf("key-%d-%03d", version-1, 0)))
	require.NoError(t, err)
	require.Nil(t, val)

	version = uint64(latestVersion)
	val, err = db.Get(storeKey1, version, []byte(fmt.Sprintf("key-%d-%03d", version-1, 0)))
	require.NoError(t, err)
	require.Equal(t, []byte(fmt.Sprintf("val-%d-%03d", version-1, 0)), val)
}
