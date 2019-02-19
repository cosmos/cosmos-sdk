package cachekv

import (
	"bytes"

	cmn "github.com/tendermint/tendermint/libs/common"
)

type pairs interface {
	get(int) cmn.KVPair
	set(int, cmn.KVPair)
	length() int
	appendAssign(cmn.KVPair)
	droplast()
}

type cmnpairs []cmn.KVPair

var _ pairs = (*cmnpairs)(nil)

func (pairs *cmnpairs) get(i int) cmn.KVPair {
	return (*pairs)[i]
}

func (pairs *cmnpairs) set(i int, pair cmn.KVPair) {
	(*pairs)[i] = pair
}

func (pairs *cmnpairs) swap(i, j int) {
	tmp := (*pairs)[i]
	(*pairs)[i] = (*pairs)[j]
	(*pairs)[j] = tmp
}

func (pairs *cmnpairs) length() int {
	return len(*pairs)
}

func (pairs *cmnpairs) droplast() {
	*pairs = (*pairs)[:len(*pairs)-1]
}

func (pairs *cmnpairs) appendAssign(pair cmn.KVPair) {
	*pairs = append((*pairs), pair)
}

// not intended to write on parent
// CONTRACT: should not call parent.{set, swap, sliceAssign}()
type cachepairs struct {
	parent pairs
	cache  []cmn.KVPair
}

var _ pairs = (*cachepairs)(nil)

func newCachePairs(parent pairs) *cachepairs {
	return &cachepairs{
		parent: parent,
		cache:  make([]cmn.KVPair, parent.length()),
	}
}

func (pairs *cachepairs) get(i int) cmn.KVPair {
	if pairs.cache[i].Key == nil {
		pairs.cache[i] = pairs.parent.get(i)
	}
	return pairs.cache[i]
}

func (pairs *cachepairs) set(i int, pair cmn.KVPair) {
	pairs.cache[i] = pair
}

func (pairs *cachepairs) length() int {
	return len(pairs.cache)
}

func (pairs *cachepairs) droplast() {
	pairs.cache = pairs.cache[:len(pairs.cache)-1]
}

func (pairs *cachepairs) appendAssign(pair cmn.KVPair) {
	panic("appendAssign should not be called on cachepairs")
}

type indexByKey map[string]int

func (ibk indexByKey) get(bz []byte) (int, bool) {
	if ibk == nil {
		panic("get() should not be called on nil indexByKey")
	}
	index, ok := ibk[string(bz)]
	return index, ok
}

func (ibk indexByKey) set(bz []byte, index int) {
	if ibk == nil {
		return
	}
	ibk[string(bz)] = index
}

func (ibk indexByKey) del(bz []byte) {
	if ibk == nil {
		return
	}
	delete(ibk, string(bz))
}

type hptr struct {
	heap  *heap
	index int
}

func (it *hptr) get() cmn.KVPair {
	return it.heap.get(it.index)
}

func (it *hptr) set(pair cmn.KVPair) {
	it.heap.set(it.index, pair)
	it.heap.indexByKey.set(pair.Key, it.index)
}

func (it *hptr) del() {
	heap := it.heap
	heap.indexByKey.del(it.get().Key)
	if it.index == heap.length()-1 {
		// no need to swap it<->last if they are same
		heap.droplast()
		return
	}
	last := heap.get(heap.length() - 1)
	it.set(last)
	heap.droplast()

	/*
		if it.heap.isEmpty() {
			return
		}
	*/

	if !it.parent().isParent(it) {
		it.siftUp()
		return
	}

	it.siftDown()
}

func (it *hptr) key() []byte {
	return it.get().Key
}

func (it *hptr) value() []byte {
	return it.get().Value
}

func (it *hptr) exists() bool {
	if it.index >= it.heap.length() {
		return false
	}
	return true
}

func (it *hptr) hasChild() bool {
	if it.index*2+1 >= it.heap.length() {
		return false
	}
	return true
}

func (it *hptr) parent() *hptr {
	return &hptr{
		heap:  it.heap,
		index: (it.index - 1) / 2,
	}
}

func (it *hptr) leftChild() *hptr {
	return &hptr{
		heap:  it.heap,
		index: it.index*2 + 1,
	}
}

func (it *hptr) rightChild() *hptr {
	return &hptr{
		heap:  it.heap,
		index: it.index*2 + 2,
	}
}

func (this *hptr) swap(that *hptr) {
	tmp := this.get()
	this.set(that.get())
	that.set(tmp)
}

func (parent *hptr) isParent(child *hptr) bool {
	comp := bytes.Compare(parent.key(), child.key())
	if parent.heap.ascending {
		return comp < 0 // parent should be smaller than child
	}
	return comp > 0 // parent should be bigger than child
}

func (it *hptr) siftUp() {
	if it.index == 0 {
		return
	}

	parent := it.parent()
	if !parent.isParent(it) {
		parent.swap(it)
		parent.siftUp()
	}
}

func (it *hptr) leafSearch() *hptr {
	left := it.leftChild()
	right := it.rightChild()

	if left.exists() {
		if right.exists() {
			if right.isParent(left) {
				return right.leafSearch()
			}
		}
		return left.leafSearch()
	}

	return it
}

func (it *hptr) siftDown() {
	grandChild := it.leafSearch()
	for it.isParent(grandChild) {
		grandChild = grandChild.parent()
	}

	bubble := grandChild.get()
	grandChild.set(it.get())
	for it.index < grandChild.index {
		grandChild = grandChild.parent()
		tmp := grandChild.get()
		grandChild.set(bubble)
		bubble = tmp
	}
}

type heap struct {
	pairs
	indexByKey indexByKey
	ascending  bool
}

func newHeapFromCache(cache map[string]cValue, ascending bool) (res *heap) {
	pairs := cmnpairs(make([]cmn.KVPair, 0, len(cache)))

	for k, cv := range cache {
		if !cv.dirty {
			continue
		}

		pairs = append(pairs, cmn.KVPair{
			Key:   []byte(k),
			Value: cv.value,
		})
	}

	return newHeap(pairs, ascending)
}

func newHeap(pairs cmnpairs, ascending bool) (res *heap) {
	res = &heap{
		pairs:      &pairs,
		indexByKey: indexByKey(make(map[string]int)),
		ascending:  ascending,
	}

	if pairs.length() == 0 {
		return
	}

	for i := 0; i < pairs.length(); i++ {
		res.indexByKey.set(pairs.get(i).Key, i)
	}
	for i := len(pairs) / 2; i >= 0; i-- {
		res.ptr(i).siftDown()
	}
	return
}

func (parent *heap) cache() (res *heap) {
	if parent == nil {
		return nil
	}
	return &heap{
		pairs: newCachePairs(parent.pairs),
		// indexByKey is for updating pairs efficiently,
		// no updating in cache heap, thus nil
		indexByKey: nil,
		ascending:  parent.ascending,
	}
}

func (heap *heap) ptr(i int) *hptr {
	return &hptr{
		heap:  heap,
		index: i,
	}
}

func (heap *heap) isEmpty() bool {
	return heap.length() == 0
}

func (heap *heap) push(pair cmn.KVPair) {
	if index, ok := heap.indexByKey.get(pair.Key); ok {
		heap.ptr(index).set(pair)
		return
	}
	heap.appendAssign(pair)
	heap.indexByKey.set(pair.Key, heap.length()-1)
	heap.ptr(heap.length() - 1).siftUp()
}

func (heap *heap) peek() cmn.KVPair {
	return heap.get(0)
}

func (heap *heap) pop() {
	root := heap.ptr(0)
	heap.indexByKey.del(root.get().Key)
	last := heap.get(heap.length() - 1)
	root.set(last)
	heap.droplast()
	if heap.isEmpty() {
		return
	}
	heap.ptr(0).siftDown()
}
