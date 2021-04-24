package exported_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/internal/proofs"
)

func TestCreateNonMembership(t *testing.T) {
	cases := map[string]struct {
		size int
		loc  proofs.Where
	}{
		"small left":   {size: 100, loc: proofs.Left},
		"small middle": {size: 100, loc: proofs.Middle},
		"small right":  {size: 100, loc: proofs.Right},
		"big left":     {size: 5431, loc: proofs.Left},
		"big middle":   {size: 5431, loc: proofs.Middle},
		"big right":    {size: 5431, loc: proofs.Right},
	}

	i := 0
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			i++
			data := proofs.BuildMap(tc.size)
			allkeys := proofs.SortedKeys(data)
			key := proofs.GetNonKey(allkeys, tc.loc)

			type serialize struct {
				Data map[string][]byte
				Key  string
			}
			sz := &serialize{Data: data, Key: key}
			blob, err := json.Marshal(sz)
			if err != nil {
				t.Fatal(err)
			}
			filename := fmt.Sprintf("%d.txt", i)
			if err := ioutil.WriteFile(filename, blob, 0755); err != nil {
				t.Fatal(err)
			}
		})
	}
}
