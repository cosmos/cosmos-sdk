package iavlx

type ProofInnerNode struct {
	Height  uint8  `json:"height"`
	Size    int64  `json:"size"`
	Version uint64 `json:"version"`
	Left    []byte `json:"left"`
	Right   []byte `json:"right"`
}

type PathToLeaf []ProofInnerNode
