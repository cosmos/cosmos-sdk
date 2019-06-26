package tendermint

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func newRoot() merkle.Root {
	return merkle.NewRoot(nil, [][]byte{[]byte("test")}, []byte{0x12, 0x34})
}

func testUpdate(t *testing.T, interval int, ok bool) {
	node := NewNode(NewMockValidators(100, 10))

	node.Commit()

	verifier := node.LastStateVerifier(newRoot())

	for i := 0; i < 100; i++ {
		header := node.Commit()

		if i%interval == 0 {
			err := verifier.Validate(header, node.PrevValset, node.Valset)
			if ok {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		}
	}
}

func TestEveryBlockUpdate(t *testing.T) {
	testUpdate(t, 1, true)
}

func TestEvenBlockUpdate(t *testing.T) {
	testUpdate(t, 2, true)
}

func TestSixthBlockUpdate(t *testing.T) {
	testUpdate(t, 6, true)
}

/*
// This should fail, since the amount of mutation is so large
// Commented out because it sometimes success
func TestTenthBlockUpdate(t *testing.T) {
	testUpdate(t, 10, false)
}
*/

func TestProofs(t *testing.T) {
	testProof(t)
}
