package a_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	_ "github.com/cosmos/gogoproto/types" // Import gogo types
	_ "google.golang.org/protobuf/types/known/wrapperspb"
)

func TestA(t *testing.T) {
	r, err := proto.MergedRegistry()

	fmt.Println(r.NumFiles())
	require.NoError(t, err)
}
