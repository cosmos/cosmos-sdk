package createnonmembershipproof

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/store/exported"
)

type serialize struct {
	Data map[string][]byte
	Key  string
}

func Fuzz(data []byte) int {
	sz := new(serialize)
	if err := json.Unmarshal(data, sz); err != nil {
		return -1
	}
	if len(sz.Data) == 0 || len(sz.Key) == 0 {
		return -1
	}
	icp, err := exported.CreateNonMembershipProof(sz.Data, []byte(sz.Key))
	if err != nil {
		return -1
	}
	if icp == nil {
		panic("nil CommitmentProof with nil error")
	}
	return 1
}
