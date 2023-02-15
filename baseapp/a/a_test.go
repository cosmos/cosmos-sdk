package a_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	_ "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// TestA is a temporary test to test file descriptors mismatch by importing
// them 1 by 1.
func TestA(t *testing.T) {
	r, err := proto.MergedRegistry()

	fmt.Println(r.NumFiles())
	require.NoError(t, err)
}
