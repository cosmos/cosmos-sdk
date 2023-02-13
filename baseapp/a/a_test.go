package a_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	// _ "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	_ "github.com/cosmos/gogoproto/types" // Import gogo types
)

func TestA(t *testing.T) {
	r, err := proto.MergedRegistry()

	fmt.Println(r.NumFiles())
	require.NoError(t, err)
}
