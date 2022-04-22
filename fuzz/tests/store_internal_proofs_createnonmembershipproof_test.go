//go:build gofuzz || go1.18

package tests

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/exported"
)

type serialize struct {
	Data map[string][]byte
	Key  string
}

func FuzzStoreInternalProofsCreateNonmembershipProof(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		sz := new(serialize)
		if err := json.Unmarshal(data, sz); err != nil {
			return
		}
		if len(sz.Data) == 0 || len(sz.Key) == 0 {
			return
		}
		icp, err := exported.CreateNonMembershipProof(sz.Data, []byte(sz.Key))
		if err != nil {
			return
		}
		if icp == nil {
			panic("nil CommitmentProof with nil error")
		}
	})
}
