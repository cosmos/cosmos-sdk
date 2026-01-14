package internal

type HashScheduler interface {
	ComputeHashes(*MemNode, *MemNode) ([]byte, []byte, error)
}

type AsyncHashScheduler struct {
	semaphore chan struct{}
}

func NewAsyncHashScheduler(maxConcurrency int32) *AsyncHashScheduler {
	return &AsyncHashScheduler{
		semaphore: make(chan struct{}, maxConcurrency),
	}
}

func (a *AsyncHashScheduler) ComputeHashes(left *MemNode, right *MemNode) (leftHash []byte, rightHash []byte, err error) {
	if left.Height() >= 4 && right.Height() >= 4 {
		select {
		case a.semaphore <- struct{}{}:
			// acquired semaphore, proceed with async computation
			leftDone := make(chan error, 1)
			go func() {
				defer func() { <-a.semaphore }() // release semaphore
				var err error                    // don't write to the outer err
				leftHash, err = left.ComputeHash(a)
				leftDone <- err
			}()
			rightHash, err = right.ComputeHash(a)
			if err != nil {
				return nil, nil, err
			}
			err = <-leftDone
			if err != nil {
				return nil, nil, err
			}
			return leftHash, rightHash, nil
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
