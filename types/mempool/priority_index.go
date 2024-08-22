package mempool

import (
	"sync"

	"github.com/huandu/skiplist"
)

// ConcurrentListElement represents a node in a concurrent priority index,
// encapsulating a skiplist element and a read-write mutex for safe concurrent access.
type ConcurrentListElement struct {
	*skiplist.Element
	mutex *sync.RWMutex
	Value interface{}
}

// Next safely retrieves the next node in the priority index.
// It acquires a read lock before accessing the next element and releases it afterward.
func (n *ConcurrentListElement) Next() *ConcurrentListElement {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	ele := n.Element.Next()
	if ele == nil {
		return nil
	}
	return &ConcurrentListElement{
		Element: ele,
		mutex:   n.mutex,
		Value:   ele.Value,
	}
}

type scoreKey struct {
	nonce  uint64
	sender string
}

type score[C comparable] struct {
	Priority C
	Weight   C
}

// ConcurrentSkipList represents a concurrent priority index,
// containing a skiplist and a map to track priority counts.
type ConcurrentSkipList[C comparable] struct {
	mutex          sync.RWMutex
	list           *skiplist.SkipList
	priorityCounts map[C]int
	scores         map[scoreKey]score[C]
}

// newConcurrentPriorityIndex initializes a new ConcurrentPriorityIndex.
// It accepts a Comparable for the skiplist and a boolean to determine if counts should be tracked.
func newConcurrentPriorityIndex[C comparable](
	listComparable skiplist.Comparable,
	priority bool,
) *ConcurrentSkipList[C] {
	i := &ConcurrentSkipList[C]{
		list: skiplist.New(listComparable),
	}
	if priority {
		i.priorityCounts = make(map[C]int)
		i.scores = make(map[scoreKey]score[C])
	}
	return i
}

// Len returns the number of elements in the priority index.
// It locks the list for reading to ensure safe access.
func (i *ConcurrentSkipList[C]) Len() int {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.list.Len()
}

// Front retrieves the first node in the priority index.
// It locks the list for reading to ensure safe access.
func (i *ConcurrentSkipList[C]) Front() *ConcurrentListElement {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	ele := i.list.Front()
	if ele == nil {
		return nil
	}
	return &ConcurrentListElement{
		Element: ele,
		mutex:   &i.mutex,
		Value:   ele.Value,
	}
}

// GetCount retrieves the count of a specific key from the priority counts map.
// It locks priorityCounts for reading to ensure safe access.
func (i *ConcurrentSkipList[C]) GetCount(key C) int {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	if i.priorityCounts == nil {
		return -1
	}
	return i.priorityCounts[key]
}

// CloneCounts creates a copy of the priority counts map.
// It locks priorityCounts for reading to ensure safe access.
func (i *ConcurrentSkipList[C]) CloneCounts() map[C]int {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	if i.priorityCounts == nil {
		return nil
	}

	counts := make(map[C]int)
	for k, v := range i.priorityCounts {
		counts[k] = v
	}
	return counts
}

// GetScore retrieves the score associated with a specific nonce and sender.
// It returns a pointer to the score if found, or nil if not found.
// It locks the scores for reading to ensure safe access.
func (i *ConcurrentSkipList[C]) GetScore(nonce uint64, sender string) *score[C] { //revive:disable:unexported-return
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	score, ok := i.scores[scoreKey{nonce: nonce, sender: sender}]
	if !ok {
		return nil
	}
	return &score
}

// Get retrieves a node corresponding to a specific key from the priority index.
// It locks the list for reading to ensure safe access.
func (i *ConcurrentSkipList[C]) Get(key txMeta[C]) *ConcurrentListElement {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	ele := i.list.Get(key)
	if ele == nil {
		return nil
	}
	return &ConcurrentListElement{
		Element: ele,
		mutex:   &i.mutex,
		Value:   ele.Value,
	}
}

// Set inserts or updates a node in the priority index with the given key and value.
// It locks priorityCounts, scores and list for writing to ensure safe access and updates the priority count.
func (i *ConcurrentSkipList[C]) Set(key txMeta[C], value any) *ConcurrentListElement {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.priorityCounts != nil {
		i.priorityCounts[key.priority]++
	}
	if i.scores != nil {
		i.scores[scoreKey{
			nonce:  key.nonce,
			sender: key.sender,
		}] = score[C]{
			Priority: key.priority,
			Weight:   key.weight,
		}
	}
	ele := i.list.Set(key, value)
	if ele == nil {
		return nil
	}
	return &ConcurrentListElement{
		Element: ele,
		mutex:   &i.mutex,
		Value:   ele.Value,
	}
}

// Remove deletes a node from the priority index using the specified key.
// It locks priorityCounts and scores for writing to ensure safe access and decrements the priority count.
func (i *ConcurrentSkipList[C]) Remove(key txMeta[C]) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.priorityCounts != nil {
		i.priorityCounts[key.priority]--
	}
	if i.scores != nil {
		delete(i.scores, scoreKey{nonce: key.nonce, sender: key.sender})
	}
	i.list.Remove(key)
}
