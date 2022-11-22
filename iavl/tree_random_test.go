package iavl

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/iavl/fastnode"
)

func TestRandomOperations(t *testing.T) {
	// In short mode (specifically, when running in CI with the race detector),
	// we only run the first couple of seeds.
	seeds := []int64{
		498727689,
		756509998,
		480459882,
		324736440,
		581827344,
		470870060,
		390970079,
		846023066,
		518638291,
		957382170,
	}

	for i, seed := range seeds {
		i, seed := i, seed
		t.Run(fmt.Sprintf("Seed %v", seed), func(t *testing.T) {
			if testing.Short() && i >= 2 {
				t.Skip("Skipping seed in short mode")
			}
			t.Parallel() // comment out to disable parallel tests, or use -parallel 1
			testRandomOperations(t, seed)
		})
	}
}

// Randomized test that runs all sorts of random operations, mirrors them in a known-good
// map, and verifies the state of the tree against the map.
func testRandomOperations(t *testing.T, randSeed int64) {
	const (
		keySize   = 16 // before base64-encoding
		valueSize = 16 // before base64-encoding

		versions          = 32   // number of final versions to generate
		reloadChance      = 0.1  // chance of tree reload after save
		deleteChance      = 0.2  // chance of random version deletion after save
		deleteRangeChance = 0.3  // chance of deleting a version range (DeleteVersionsRange)
		deleteMultiChance = 0.3  // chance of deleting multiple versions (DeleteVersions)
		deleteMax         = 5    // max number of versions to delete
		revertChance      = 0.05 // chance to revert tree to random version with LoadVersionForOverwriting
		syncChance        = 0.2  // chance of enabling sync writes on tree load
		cacheChance       = 0.4  // chance of enabling caching
		cacheSizeMax      = 256  // maximum size of cache (will be random from 1)

		versionOps  = 64  // number of operations (create/update/delete) per version
		updateRatio = 0.4 // ratio of updates out of all operations
		deleteRatio = 0.2 // ratio of deletes out of all operations
	)

	r := rand.New(rand.NewSource(randSeed))

	// loadTree loads the last persisted version of a tree with random pruning settings.
	loadTree := func(levelDB db.DB) (tree *MutableTree, version int64, options *Options) {
		var err error
		options = &Options{
			Sync: r.Float64() < syncChance,
		}
		// set the cache size regardless of whether caching is enabled. This ensures we always
		// call the RNG the same number of times, such that changing settings does not affect
		// the RNG sequence.
		cacheSize := int(r.Int63n(cacheSizeMax + 1))
		if !(r.Float64() < cacheChance) {
			cacheSize = 0
		}
		tree, err = NewMutableTreeWithOpts(levelDB, cacheSize, options, false)
		require.NoError(t, err)
		version, err = tree.Load()
		require.NoError(t, err)
		t.Logf("Loaded version %v (sync=%v cache=%v)", version, options.Sync, cacheSize)
		return
	}

	// generates random keys and values
	randString := func(size int) string {
		buf := make([]byte, size)
		r.Read(buf)
		return base64.StdEncoding.EncodeToString(buf)
	}

	// Use the same on-disk database for the entire run.
	tempdir, err := ioutil.TempDir("", "iavl")
	require.NoError(t, err)
	defer os.RemoveAll(tempdir)

	levelDB, err := db.NewGoLevelDB("leveldb", tempdir)
	require.NoError(t, err)

	tree, version, _ := loadTree(levelDB)

	// Set up a mirror of the current IAVL state, as well as the history of saved mirrors
	// on disk and in memory. Since pruning was removed we currently persist all versions,
	// thus memMirrors is never used, but it is left here for the future when it is re-introduces.
	mirror := make(map[string]string, versionOps)
	mirrorKeys := make([]string, 0, versionOps)
	diskMirrors := make(map[int64]map[string]string)
	memMirrors := make(map[int64]map[string]string)

	for version < versions {
		for i := 0; i < versionOps; i++ {
			switch {
			case len(mirror) > 0 && r.Float64() < deleteRatio:
				index := r.Intn(len(mirrorKeys))
				key := mirrorKeys[index]
				mirrorKeys = append(mirrorKeys[:index], mirrorKeys[index+1:]...)
				_, removed, err := tree.Remove([]byte(key))
				require.NoError(t, err)
				require.True(t, removed)
				delete(mirror, key)

			case len(mirror) > 0 && r.Float64() < updateRatio:
				key := mirrorKeys[r.Intn(len(mirrorKeys))]
				value := randString(valueSize)
				updated, err := tree.Set([]byte(key), []byte(value))
				require.NoError(t, err)
				require.True(t, updated)
				mirror[key] = value

			default:
				key := randString(keySize)
				value := randString(valueSize)
				for has, err := tree.Has([]byte(key)); has && err == nil; {
					key = randString(keySize)
				}
				updated, err := tree.Set([]byte(key), []byte(value))
				require.NoError(t, err)
				require.False(t, updated)
				mirror[key] = value
				mirrorKeys = append(mirrorKeys, key)
			}
		}
		_, version, err = tree.SaveVersion()
		require.NoError(t, err)

		t.Logf("Saved tree at version %v with %v keys and %v versions",
			version, tree.Size(), len(tree.AvailableVersions()))

		// Verify that the version matches the mirror.
		assertMirror(t, tree, mirror, 0)

		// Save the mirror as a disk mirror, since we currently persist all versions.
		diskMirrors[version] = copyMirror(mirror)

		// Delete random versions if requested, but never the latest version.
		if r.Float64() < deleteChance {
			versions := getMirrorVersions(diskMirrors, memMirrors)
			switch {
			case len(versions) < 2:

			case r.Float64() < deleteRangeChance:
				indexFrom := r.Intn(len(versions) - 1)
				from := versions[indexFrom]
				batch := r.Intn(deleteMax)
				if batch > len(versions[indexFrom:])-2 {
					batch = len(versions[indexFrom:]) - 2
				}
				to := versions[indexFrom+batch] + 1
				t.Logf("Deleting versions %v-%v", from, to-1)
				err = tree.DeleteVersionsRange(int64(from), int64(to))
				require.NoError(t, err)
				for version := from; version < to; version++ {
					delete(diskMirrors, int64(version))
					delete(memMirrors, int64(version))
				}

			// adjust probability to take into account probability of range delete not happening
			case r.Float64() < deleteMultiChance/(1.0-deleteRangeChance):
				deleteVersions := []int64{}
				desc := ""
				batchSize := 1 + r.Intn(deleteMax)
				if batchSize > len(versions)-1 {
					batchSize = len(versions) - 1
				}
				for _, i := range r.Perm(len(versions) - 1)[:batchSize] {
					deleteVersions = append(deleteVersions, int64(versions[i]))
					delete(diskMirrors, int64(versions[i]))
					delete(memMirrors, int64(versions[i]))
					if len(desc) > 0 {
						desc += ","
					}
					desc += fmt.Sprintf("%v", versions[i])
				}
				t.Logf("Deleting versions %v", desc)
				err = tree.DeleteVersions(deleteVersions...)
				require.NoError(t, err)

			default:
				i := r.Intn(len(versions) - 1)
				deleteVersion := int64(versions[i])
				t.Logf("Deleting version %v", deleteVersion)
				err = tree.DeleteVersion(deleteVersion)
				require.NoError(t, err)
				delete(diskMirrors, deleteVersion)
				delete(memMirrors, deleteVersion)
			}
		}

		// Reload tree from last persisted version if requested, checking that it matches the
		// latest disk mirror version and discarding memory mirrors.
		if r.Float64() < reloadChance {
			tree, version, _ = loadTree(levelDB)
			assertMaxVersion(t, tree, version, diskMirrors)
			memMirrors = make(map[int64]map[string]string)
			mirror = copyMirror(diskMirrors[version])
			mirrorKeys = getMirrorKeys(mirror)
		}

		// Revert tree to historical version if requested, deleting all subsequent versions.
		if r.Float64() < revertChance {
			versions := getMirrorVersions(diskMirrors, memMirrors)
			if len(versions) > 1 {
				version = int64(versions[r.Intn(len(versions)-1)])
				t.Logf("Reverting to version %v", version)
				_, err = tree.LoadVersionForOverwriting(version)
				require.NoError(t, err, "Failed to revert to version %v", version)
				if m, ok := diskMirrors[version]; ok {
					mirror = copyMirror(m)
				} else if m, ok := memMirrors[version]; ok {
					mirror = copyMirror(m)
				} else {
					t.Fatalf("Mirror not found for revert target %v", version)
				}
				mirrorKeys = getMirrorKeys(mirror)
				for v := range diskMirrors {
					if v > version {
						delete(diskMirrors, v)
					}
				}
				for v := range memMirrors {
					if v > version {
						delete(memMirrors, v)
					}
				}
			}
		}

		// Verify all historical versions.
		assertVersions(t, tree, diskMirrors, memMirrors)

		for diskVersion, diskMirror := range diskMirrors {
			assertMirror(t, tree, diskMirror, diskVersion)
		}

		for memVersion, memMirror := range memMirrors {
			assertMirror(t, tree, memMirror, memVersion)
		}
	}

	// Once we're done, delete all prior versions in random order, make sure all orphans have been
	// removed, and check that the latest versions matches the mirror.
	remaining := tree.AvailableVersions()
	remaining = remaining[:len(remaining)-1]

	switch {
	case len(remaining) == 0:

	case r.Float64() < deleteRangeChance:
		t.Logf("Deleting versions %v-%v", remaining[0], remaining[len(remaining)-1])
		err = tree.DeleteVersionsRange(int64(remaining[0]), int64(remaining[len(remaining)-1]+1))
		require.NoError(t, err)

	// adjust probability to take into account probability of range delete not happening
	case r.Float64() < deleteMultiChance/(1.0-deleteRangeChance):
		deleteVersions := []int64{}
		desc := ""
		for _, i := range r.Perm(len(remaining)) {
			deleteVersions = append(deleteVersions, int64(remaining[i]))
			if len(desc) > 0 {
				desc += ","
			}
			desc += fmt.Sprintf("%v", remaining[i])
		}
		t.Logf("Deleting versions %v", desc)
		err = tree.DeleteVersions(deleteVersions...)
		require.NoError(t, err)

	default:
		for len(remaining) > 0 {
			i := r.Intn(len(remaining))
			deleteVersion := int64(remaining[i])
			remaining = append(remaining[:i], remaining[i+1:]...)
			t.Logf("Deleting version %v", deleteVersion)
			err = tree.DeleteVersion(deleteVersion)
			require.NoError(t, err)
		}
	}

	require.EqualValues(t, []int{int(version)}, tree.AvailableVersions())
	assertMirror(t, tree, mirror, version)
	assertMirror(t, tree, mirror, 0)
	assertOrphans(t, tree, 0)
	t.Logf("Final version %v is correct, with no stray orphans", version)

	// Now, let's delete all remaining key/value pairs, and make sure no stray
	// data is left behind in the database.
	prevVersion := tree.Version()
	keys := [][]byte{}
	tree.Iterate(func(key, value []byte) bool {
		keys = append(keys, key)
		return false
	})
	for _, key := range keys {
		_, removed, err := tree.Remove(key)
		require.NoError(t, err)
		require.True(t, removed)
	}
	_, _, err = tree.SaveVersion()
	require.NoError(t, err)
	err = tree.DeleteVersion(prevVersion)
	require.NoError(t, err)
	assertEmptyDatabase(t, tree)
	t.Logf("Final version %v deleted, no stray database entries", prevVersion)
}

// Checks that the database is empty, only containing a single root entry
// at the given version.
func assertEmptyDatabase(t *testing.T, tree *MutableTree) {
	version := tree.Version()
	iter, err := tree.ndb.db.Iterator(nil, nil)
	require.NoError(t, err)

	var foundKeys []string
	for ; iter.Valid(); iter.Next() {
		foundKeys = append(foundKeys, string(iter.Key()))
	}
	require.NoError(t, iter.Error())
	require.EqualValues(t, 2, len(foundKeys), "Found %v database entries, expected 1", len(foundKeys)) // 1 for storage version and 1 for root

	firstKey := foundKeys[0]
	secondKey := foundKeys[1]

	require.True(t, strings.HasPrefix(firstKey, metadataKeyFormat.Prefix()))
	require.True(t, strings.HasPrefix(secondKey, rootKeyFormat.Prefix()))

	require.Equal(t, string(metadataKeyFormat.KeyBytes([]byte(storageVersionKey))), firstKey, "Unexpected storage version key")

	storageVersionValue, err := tree.ndb.db.Get([]byte(firstKey))
	require.NoError(t, err)
	latestVersion, err := tree.ndb.getLatestVersion()
	require.NoError(t, err)
	require.Equal(t, fastStorageVersionValue+fastStorageVersionDelimiter+strconv.Itoa(int(latestVersion)), string(storageVersionValue))

	var foundVersion int64
	rootKeyFormat.Scan([]byte(secondKey), &foundVersion)
	require.Equal(t, version, foundVersion, "Unexpected root version")
}

// Checks that the tree has the given number of orphan nodes.
func assertOrphans(t *testing.T, tree *MutableTree, expected int) {
	count := 0
	err := tree.ndb.traverseOrphans(func(k, v []byte) error {
		count++
		return nil
	})
	require.Nil(t, err)
	require.EqualValues(t, expected, count, "Expected %v orphans, got %v", expected, count)
}

// Checks that a version is the maximum mirrored version.
func assertMaxVersion(t *testing.T, tree *MutableTree, version int64, mirrors map[int64]map[string]string) {
	max := int64(0)
	for v := range mirrors {
		if v > max {
			max = v
		}
	}
	require.Equal(t, max, version)
}

// Checks that a mirror, optionally for a given version, matches the tree contents.
func assertMirror(t *testing.T, tree *MutableTree, mirror map[string]string, version int64) {
	var err error
	itree := tree.ImmutableTree
	if version > 0 {
		itree, err = tree.GetImmutable(version)
		require.NoError(t, err, "loading version %v", version)
	}
	// We check both ways: first check that iterated keys match the mirror, then iterate over the
	// mirror and check with get. This is to exercise both the iteration and Get() code paths.
	iterated := 0
	itree.Iterate(func(key, value []byte) bool {
		require.Equal(t, string(value), mirror[string(key)], "Invalid value for key %q", key)
		iterated++
		return false
	})
	require.EqualValues(t, len(mirror), itree.Size())
	require.EqualValues(t, len(mirror), iterated)
	for key, value := range mirror {
		actualFast, err := itree.Get([]byte(key))
		require.NoError(t, err)
		require.Equal(t, value, string(actualFast))
		_, actual, err := itree.GetWithIndex([]byte(key))
		require.NoError(t, err)
		require.Equal(t, value, string(actual))
	}

	assertFastNodeCacheIsLive(t, tree, mirror, version)
	assertFastNodeDiskIsLive(t, tree, mirror, version)
}

// Checks that fast node cache matches live state.
func assertFastNodeCacheIsLive(t *testing.T, tree *MutableTree, mirror map[string]string, version int64) {
	latestVersion, err := tree.ndb.getLatestVersion()
	require.NoError(t, err)
	if latestVersion != version {
		// The fast node cache check should only be done to the latest version
		return
	}

	require.Equal(t, len(mirror), tree.ndb.fastNodeCache.Len())
	for k, v := range mirror {
		require.True(t, tree.ndb.fastNodeCache.Has([]byte(k)), "cached fast node must be in live tree")
		mirrorNode := tree.ndb.fastNodeCache.Get([]byte(k))
		require.Equal(t, []byte(v), mirrorNode.(*fastnode.Node).GetValue(), "cached fast node's value must be equal to live state value")
	}
}

// Checks that fast nodes on disk match live state.
func assertFastNodeDiskIsLive(t *testing.T, tree *MutableTree, mirror map[string]string, version int64) {
	latestVersion, err := tree.ndb.getLatestVersion()
	require.NoError(t, err)
	if latestVersion != version {
		// The fast node disk check should only be done to the latest version
		return
	}

	count := 0
	err = tree.ndb.traverseFastNodes(func(keyWithPrefix, v []byte) error {
		key := keyWithPrefix[1:]
		count++
		fastNode, err := fastnode.DeserializeNode(key, v)
		require.Nil(t, err)

		mirrorVal := mirror[string(fastNode.GetKey())]

		require.NotNil(t, mirrorVal)
		require.Equal(t, []byte(mirrorVal), fastNode.GetValue())
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, len(mirror), count)
}

// Checks that all versions in the tree are present in the mirrors, and vice-versa.
func assertVersions(t *testing.T, tree *MutableTree, mirrors ...map[int64]map[string]string) {
	require.Equal(t, getMirrorVersions(mirrors...), tree.AvailableVersions())
}

// copyMirror copies a mirror map.
func copyMirror(mirror map[string]string) map[string]string {
	c := make(map[string]string, len(mirror))
	for k, v := range mirror {
		c[k] = v
	}
	return c
}

// getMirrorKeys returns the keys of a mirror, unsorted.
func getMirrorKeys(mirror map[string]string) []string {
	keys := make([]string, 0, len(mirror))
	for key := range mirror {
		keys = append(keys, key)
	}
	return keys
}

// getMirrorVersions returns the versions of the given mirrors, sorted. Returns []int to
// match tree.AvailableVersions().
func getMirrorVersions(mirrors ...map[int64]map[string]string) []int {
	versionMap := make(map[int]bool)
	for _, m := range mirrors {
		for version := range m {
			versionMap[int(version)] = true
		}
	}
	versions := make([]int, 0, len(versionMap))
	for version := range versionMap {
		versions = append(versions, version)
	}
	sort.Ints(versions)
	return versions
}
