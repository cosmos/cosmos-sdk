package mempool

import (
	"sync"
	"testing"

	"github.com/huandu/skiplist"
	"github.com/stretchr/testify/require"
)

func TestConcurrentPriorityNode_Next(t *testing.T) {
	list := skiplist.New(skiplist.Int64)
	mutex := &sync.RWMutex{}

	for i := 0; i < 5; i++ {
		list.Set(int64(i), int64(i))
	}

	firstEle := list.Front()
	require.NotNil(t, firstEle)
	node := &ConcurrentListElement{
		Element: firstEle,
		mutex:   mutex,
	}

	for i := 0; i < 5; i++ {
		require.NotNil(t, node)
		nextNode := node.Next()
		if nextNode != nil {
			expected := int64(i + 1)
			require.Equal(t, expected, nextNode.Value)
		}
		node = nextNode
	}

	// expected node to be nil after traversing all elements
	require.Nil(t, node)
}

func TestConcurrentPriorityIndex_Len(t *testing.T) {
	index := newConcurrentPriorityIndex[int](skiplist.LessThanFunc(func(a, b any) int {
		return skiplist.Uint64.Compare(b.(txMeta[int]).nonce, a.(txMeta[int]).nonce)
	}), true)

	total := 5
	for i := 0; i < total; i++ {
		index.Set(txMeta[int]{nonce: uint64(i)}, i)
	}

	require.Equal(t, total, index.Len())
}

func TestConcurrentPriorityIndex_Front(t *testing.T) {
	index := newConcurrentPriorityIndex[int](skiplist.LessThanFunc(func(a, b any) int {
		return skiplist.Int.Compare(b.(txMeta[int]).priority, a.(txMeta[int]).priority)
	}), true)

	for i := 0; i < 5; i++ {
		index.Set(txMeta[int]{priority: i}, i)
	}

	frontNode := index.Front()
	require.NotNil(t, frontNode)
	require.Equal(t, 0, frontNode.Element.Value)
}

func TestConcurrentSkipList_GetCount(t *testing.T) {
	list := &ConcurrentSkipList[int]{
		mutex:          sync.RWMutex{},
		priorityCounts: make(map[int]int),
	}

	list.priorityCounts = nil
	count := list.GetCount(10)
	require.Equal(t, -1, count)

	list.priorityCounts = make(map[int]int)
	list.priorityCounts[10] = 5
	list.priorityCounts[20] = 3

	count = list.GetCount(10)
	require.Equal(t, 5, count)

	count = list.GetCount(20)
	require.Equal(t, 3, count)

	count = list.GetCount(30)
	require.Equal(t, 0, count)
}

func TestConcurrentSkipList_CloneCounts(t *testing.T) {
	list := &ConcurrentSkipList[int]{
		mutex:          sync.RWMutex{},
		priorityCounts: make(map[int]int),
	}
	list.priorityCounts = nil
	counts := list.CloneCounts()
	require.Nil(t, counts)
	list.priorityCounts = map[int]int{
		10: 5,
		20: 3,
		30: 7,
	}

	counts = list.CloneCounts()
	require.NotNil(t, counts)
	require.Equal(t, 5, counts[10])
	require.Equal(t, 3, counts[20])
	require.Equal(t, 7, counts[30])

	// check the cloned map is a separate instance after modified
	counts[10] = 99
	require.Equal(t, 5, list.priorityCounts[10])
}

func TestConcurrentSkipList_GetScore(t *testing.T) {
	list := &ConcurrentSkipList[int]{
		mutex: sync.RWMutex{},
		scores: map[scoreKey]score[int]{
			{nonce: 1, sender: "sender1"}: {Priority: 10, Weight: 1},
			{nonce: 2, sender: "sender2"}: {Priority: 20, Weight: 2},
		},
	}

	score := list.GetScore(1, "sender1")
	require.NotNil(t, score)
	require.Equal(t, 10, score.Priority)
	require.Equal(t, 1, score.Weight)

	score = list.GetScore(1, "sender2")
	require.Nil(t, score)
	score = list.GetScore(2, "sender1")
	require.Nil(t, score)
	score = list.GetScore(3, "sender3")
	require.Nil(t, score)
}

func TestConcurrentSkipList_Get(t *testing.T) {
	list := skiplist.New(skiplist.LessThanFunc(func(a, b any) int {
		return skiplist.Uint64.Compare(b.(txMeta[int]).nonce, a.(txMeta[int]).nonce)
	}))
	concurrentList := &ConcurrentSkipList[int]{
		mutex: sync.RWMutex{},
		list:  list,
	}

	for i := 0; i < 5; i++ {
		key := txMeta[int]{nonce: uint64(i)}
		concurrentList.Set(key, i)
	}

	key := txMeta[int]{nonce: 2}
	ele := concurrentList.Get(key)
	require.NotNil(t, ele)
	require.Equal(t, 2, ele.Value)

	nonExist := txMeta[int]{nonce: 10}
	ele = concurrentList.Get(nonExist)
	require.Nil(t, ele)
}

func TestConcurrentSkipList_Set(t *testing.T) {
	list := skiplist.New(skiplist.LessThanFunc(func(a, b any) int {
		return skiplist.Uint64.Compare(a.(txMeta[int]).nonce, b.(txMeta[int]).nonce)
	}))

	concurrentList := &ConcurrentSkipList[int]{
		mutex:          sync.RWMutex{},
		list:           list,
		priorityCounts: make(map[int]int),
		scores:         make(map[scoreKey]score[int]),
	}

	key1 := txMeta[int]{nonce: 1, sender: "sender1", priority: 10, weight: 2}
	ele1 := concurrentList.Set(key1, 1)
	require.NotNil(t, ele1)
	require.Equal(t, 1, ele1.Element.Value)
	require.Equal(t, 1, concurrentList.priorityCounts[key1.priority])
	require.Equal(t, score[int]{Priority: 10, Weight: 2}, concurrentList.scores[scoreKey{nonce: 1, sender: "sender1"}])

	// update existing element
	key2 := txMeta[int]{nonce: 1, sender: "sender1", priority: 20, weight: 3}
	ele2 := concurrentList.Set(key2, 2)
	require.NotNil(t, ele2)
	require.Equal(t, 2, ele2.Element.Value)
	require.Equal(t, 1, concurrentList.priorityCounts[key2.priority])
	require.Equal(t, score[int]{Priority: 20, Weight: 3}, concurrentList.scores[scoreKey{nonce: 1, sender: "sender1"}])

	// inserting new element
	key3 := txMeta[int]{nonce: 2, sender: "sender2", priority: 15, weight: 1}
	ele3 := concurrentList.Set(key3, 3)
	require.NotNil(t, ele3)
	require.Equal(t, 3, ele3.Element.Value)
	require.Equal(t, 1, concurrentList.priorityCounts[key3.priority])
	require.Equal(t, score[int]{Priority: 15, Weight: 1}, concurrentList.scores[scoreKey{nonce: 2, sender: "sender2"}])
}

func TestConcurrentSkipList_Remove(t *testing.T) {
	list := skiplist.New(skiplist.LessThanFunc(func(a, b any) int {
		return skiplist.Uint64.Compare(a.(txMeta[int]).nonce, b.(txMeta[int]).nonce)
	}))

	concurrentList := &ConcurrentSkipList[int]{
		mutex:          sync.RWMutex{},
		list:           list,
		priorityCounts: make(map[int]int),
		scores:         make(map[scoreKey]score[int]),
	}

	key1 := txMeta[int]{nonce: 1, sender: "sender1", priority: 10, weight: 2}
	concurrentList.Set(key1, 1)

	key2 := txMeta[int]{nonce: 2, sender: "sender2", priority: 20, weight: 3}
	concurrentList.Set(key2, 2)

	require.Equal(t, 1, concurrentList.priorityCounts[key1.priority])
	require.Equal(t, 1, concurrentList.priorityCounts[key2.priority])
	require.Equal(t, score[int]{Priority: 10, Weight: 2}, concurrentList.scores[scoreKey{nonce: 1, sender: "sender1"}])
	require.Equal(t, score[int]{Priority: 20, Weight: 3}, concurrentList.scores[scoreKey{nonce: 2, sender: "sender2"}])

	concurrentList.Remove(key1)
	require.Equal(t, 0, concurrentList.priorityCounts[key1.priority])
	require.NotContains(t, concurrentList.scores, scoreKey{nonce: key1.nonce, sender: key1.sender})
	require.Nil(t, concurrentList.Get(key1))

	require.Equal(t, 1, concurrentList.priorityCounts[key2.priority])
	require.Equal(t, score[int]{Priority: 20, Weight: 3}, concurrentList.scores[scoreKey{nonce: 2, sender: "sender2"}])
}
