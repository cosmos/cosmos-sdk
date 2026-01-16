package iavlx

import "sync"

type NonBlockingQueue[T any] struct {
	mu     sync.Mutex
	cond   *sync.Cond
	queue  []T
	closed bool
}

func NewNonBlockingQueue[T any]() *NonBlockingQueue[T] {
	res := &NonBlockingQueue[T]{}
	res.cond = sync.NewCond(&res.mu)
	return res
}

func (q *NonBlockingQueue[T]) Send(item T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		panic("NonBlockingQueue is closed, can't send")
	}
	q.queue = append(q.queue, item)
	q.cond.Signal()
}

func (q *NonBlockingQueue[T]) Receive() []T {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.queue) == 0 && !q.closed {
		q.cond.Wait()
	}

	res := q.queue
	q.queue = nil
	return res
}

func (q *NonBlockingQueue[T]) MaybeReceive() (batch []T, closed bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.queue) == 0 && !q.closed {
		return nil, false
	}

	res := q.queue
	q.queue = nil
	return res, q.closed
}

func (q *NonBlockingQueue[T]) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.cond.Broadcast()
}
