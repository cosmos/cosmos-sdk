package internal

import "context"

type HashScheduler interface {
	ComputeHashes(*MemNode, *MemNode) ([]byte, []byte, error)
}

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

func (a *AsyncHashScheduler) ComputeHashes(left *MemNode, right *MemNode) (leftHash []byte, rightHash []byte, err error) {
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

func computeHashsSync(left *MemNode, right *MemNode, scheduler HashScheduler) (leftHash []byte, rightHash []byte, err error) {
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

func (s SyncHashScheduler) ComputeHashes(node *MemNode, node2 *MemNode) ([]byte, []byte, error) {
	return computeHashsSync(node, node2, s)
}

var _ HashScheduler = (*SyncHashScheduler)(nil)
