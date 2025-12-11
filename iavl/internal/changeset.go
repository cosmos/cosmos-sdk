package internal

import "fmt"

// NOTE: This is a placeholder implementation. We will add the implementation in a future PR.

type Changeset struct{}

func (cs *Changeset) Resolve(id NodeID, fileIdx uint32) (Node, error) {
	return nil, fmt.Errorf("not implemented")
}
