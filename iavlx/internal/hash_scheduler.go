package internal

import "context"

// HashScheduler controls how child node hashes are computed during tree hashing.
// The tree hash is computed bottom-up: each branch node's hash depends on its children's hashes.
// ComputeHashes takes a branch's left and right children and returns both hashes.
//
// Two implementations:
//   - SyncHashScheduler: computes both hashes sequentially in the current goroutine.
//     Used for leaf hash pre-computation and anywhere parallelism isn't beneficial.
//   - AsyncHashScheduler: computes left and right hashes in parallel goroutines when both
//     subtrees are tall enough (height >= 4) to justify the goroutine overhead.
//     Uses a semaphore to cap concurrency at NumCPU. Falls back to sync when the semaphore
//     is full or subtrees are too shallow.
type HashScheduler interface {
	ComputeHashes(*MemNode, *MemNode) ([]byte, []byte, error)
}

// AsyncHashScheduler parallelizes hash computation across subtrees using a bounded goroutine pool.
// It only spawns a goroutine when both children have height >= 4 (enough work to offset the
// overhead) AND the semaphore has capacity. Otherwise it falls back to synchronous computation.
type AsyncHashScheduler struct {
	semaphore chan struct{}
	ctx       context.Context
}

func NewAsyncHashScheduler(ctx context.Context, maxConcurrency int32) *AsyncHashScheduler {
	return &AsyncHashScheduler{
		ctx:       ctx,
		semaphore: make(chan struct{}, maxConcurrency),
	}
}

func (a *AsyncHashScheduler) ComputeHashes(left, right *MemNode) (leftHash, rightHash []byte, err error) {
	if err := a.ctx.Err(); err != nil {
		return nil, nil, err
	}
	if left.Height() >= 4 && right.Height() >= 4 {
		select {
		case a.semaphore <- struct{}{}:
			// acquired semaphore, proceed with async computation
			type hashResult struct {
				hash []byte
				err  error
			}
			leftDone := make(chan hashResult, 1)
			go func() {
				defer func() { <-a.semaphore }() // release semaphore
				h, err := left.ComputeHash(a)
				leftDone <- hashResult{h, err}
			}()
			rightHash, err = right.ComputeHash(a)
			if err != nil {
				<-leftDone // wait for goroutine to finish before returning
				return nil, nil, err
			}
			lr := <-leftDone
			if lr.err != nil {
				return nil, nil, lr.err
			}
			return lr.hash, rightHash, nil
		default:
			return computeHashsSync(left, right, a)
		}
	} else {
		return computeHashsSync(left, right, a)
	}
}

func computeHashsSync(left, right *MemNode, scheduler HashScheduler) (leftHash, rightHash []byte, err error) {
	leftHash, err = left.ComputeHash(scheduler)
	if err != nil {
		return nil, nil, err
	}

	rightHash, err = right.ComputeHash(scheduler)
	if err != nil {
		return nil, nil, err
	}
	return leftHash, rightHash, nil
}

var _ HashScheduler = (*AsyncHashScheduler)(nil)

type SyncHashScheduler struct{}

func (s SyncHashScheduler) ComputeHashes(node, node2 *MemNode) ([]byte, []byte, error) {
	return computeHashsSync(node, node2, s)
}

var _ HashScheduler = (*SyncHashScheduler)(nil)
