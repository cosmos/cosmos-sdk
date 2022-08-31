//go:build gofuzz || go1.18

package tests

import (
	"encoding/json"
	"testing"

	db "github.com/cosmos/cosmos-db"
	iavlproofs "github.com/cosmos/cosmos-sdk/store/tools/ics23/iavl"
	"github.com/cosmos/iavl"
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
		tree, err := iavl.NewMutableTree(db.NewMemDB(), 0)
		if err != nil {
			t.Fatal(err)
		}
		for k, v := range sz.Data {
			tree.Set([]byte(k), v)
		}
		icp, err := iavlproofs.CreateNonMembershipProof(tree, []byte(sz.Key))
		if err != nil {
			return
		}
		if icp == nil {
			panic("nil CommitmentProof with nil error")
		}
	})
}
