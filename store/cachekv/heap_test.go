package cachekv

/*
import (
	"bytes"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type ascpairs cmnpairs

var _ sort.Interface = ascpairs{}

func (pairs ascpairs) Len() int {
	return len(pairs)
}

func (pairs ascpairs) Less(i, j int) bool {
	return bytes.Compare(pairs[i].Key, pairs[j].Key) == -1
}

func (pairs ascpairs) Swap(i, j int) {
	tmp := pairs[i]
	pairs[i] = pairs[j]
	pairs[j] = tmp
}

type descpairs cmnpairs

var _ sort.Interface = descpairs{}

func (pairs descpairs) Len() int {
	return len(pairs)
}

func (pairs descpairs) Less(i, j int) bool {
	return bytes.Compare(pairs[i].Key, pairs[j].Key) == 1
}

func (pairs descpairs) Swap(i, j int) {
	tmp := pairs[i]
	pairs[i] = pairs[j]
	pairs[j] = tmp
}

func assertValidHeap(t *testing.T, it *hptr) {
	left := it.leftChild()
	if left.exists() {
		require.True(t, it.isParent(left))
		assertValidHeap(t, left)
	}

	right := it.rightChild()
	if right.exists() {
		require.True(t, it.isParent(right))
		assertValidHeap(t, right)
	}
}

func testHeap(t *testing.T, size int, ascending bool, bzgen func(i int) []byte) {
	pairs1 := make([]cmn.KVPair, 0, size)
	pairs2 := make([]cmn.KVPair, 0, size)

	for i := 0; i < size; i++ {
		bz := bzgen(i)
		pairs1 = append(pairs1, cmn.KVPair{Key: bz, Value: []byte{}})
		pairs2 = append(pairs2, cmn.KVPair{Key: bz, Value: []byte{}})
	}

	heap := newHeap(cmnpairs(pairs1), ascending)
	assertValidHeap(t, heap.ptr(0))
	if ascending {
		sort.Sort(ascpairs(pairs2))
	} else {
		sort.Sort(descpairs(pairs2))
	}
	for _, pair := range pairs2 {
		require.Equal(t, pair.Key, heap.peek().Key)
		heap.pop()
	}

}

func TestAscendingHeapReversedElements(t *testing.T) {
	testHeap(t, 20, true, func(i int) []byte { return []byte{byte(255 - i)} })
}

func TestDescendingHeapReversedElements(t *testing.T) {
	testHeap(t, 20, false, func(i int) []byte { return []byte{byte(i)} })
}

func randgen(_ int) (res []byte) {
	res = make([]byte, 4)
	rand.Read(res)
	return
}

func TestAscendingHeapRandomElements(t *testing.T) {
	testHeap(t, 100000, true, randgen)
}

func TestDescendingHeapRandomElements(t *testing.T) {
	testHeap(t, 100000, false, randgen)
}

func TestHeapDuplicateElements(t *testing.T) {
	size := 1000

	pairs1 := make([]cmn.KVPair, 0, size)
	pairs2 := make([]cmn.KVPair, 0, size)

	for i := 0; i < size; i++ {
		bz := randgen(i)
		pairs1 = append(pairs1, cmn.KVPair{Key: bz, Value: bz})
		pairs2 = append(pairs2, cmn.KVPair{Key: bz, Value: bz})
	}

	heap := newHeap(cmnpairs(pairs1), true)
	for _, pair := range pairs2 {
		heap.push(pair)
	}

	require.Equal(t, size, heap.length())
}

func TestHeapDeleteElements(t *testing.T) {
	size := 1000

	pairs1 := make([]cmn.KVPair, 0, size)
	pairs2 := make([]cmn.KVPair, 0, size)

	for i := 0; i < size; i++ {
		bz := randgen(i)
		pairs1 = append(pairs1, cmn.KVPair{Key: bz, Value: bz})
		pairs2 = append(pairs2, cmn.KVPair{Key: bz, Value: bz})
	}

	heap := newHeap(cmnpairs(pairs1), true)
	for _, pair := range pairs2 {
		heap.del(pair.Key)
	}

	require.Equal(t, 0, heap.length())
}
*/
