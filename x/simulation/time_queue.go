package simulation

import (
	"container/heap"
	"time"

	"github.com/cosmos/cosmos-sdk/types/simulation"
)

// timeOpQueue is a min-heap of FutureOperation ordered by BlockTime.
type timeOpQueue struct {
	items []simulation.FutureOperation
}

// newTimeOpQueue creates an empty time operation queue.
func newTimeOpQueue() *timeOpQueue {
	q := &timeOpQueue{items: make([]simulation.FutureOperation, 0)}
	heap.Init(q)
	return q
}

// Len implements heap.Interface.
func (q *timeOpQueue) Len() int { return len(q.items) }

// Less implements heap.Interface.
func (q *timeOpQueue) Less(i, j int) bool {
	return q.items[i].BlockTime.Before(q.items[j].BlockTime)
}

// Swap implements heap.Interface.
func (q *timeOpQueue) Swap(i, j int) { q.items[i], q.items[j] = q.items[j], q.items[i] }

// Push implements heap.Interface.
func (q *timeOpQueue) Push(x any) {
	q.items = append(q.items, x.(simulation.FutureOperation))
}

// Pop implements heap.Interface.
func (q *timeOpQueue) Pop() any {
	n := len(q.items)
	x := q.items[n-1]
	q.items = q.items[:n-1]
	return x
}

// Peek returns the earliest FutureOperation without removing it.
func (q *timeOpQueue) Peek() (simulation.FutureOperation, bool) {
	if len(q.items) == 0 {
		return simulation.FutureOperation{}, false
	}
	return q.items[0], true
}

// PopDue pops and returns all operations with BlockTime strictly before currentTime.
func (q *timeOpQueue) PopDue(currentTime time.Time) []simulation.FutureOperation {
	due := make([]simulation.FutureOperation, 0)
	for q.Len() > 0 {
		top, _ := q.Peek()
		if !currentTime.After(top.BlockTime) {
			break
		}
		op := heap.Pop(q).(simulation.FutureOperation)
		due = append(due, op)
	}
	return due
}
