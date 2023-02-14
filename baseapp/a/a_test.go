package a_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	_ "github.com/cosmos/gogoproto/protoc-gen-gogo/descriptor"
	_ "github.com/cosmos/gogoproto/types" // Import gogo types
	_ "google.golang.org/protobuf/types/descriptorpb"
	_ "google.golang.org/protobuf/types/known/anypb"
	_ "google.golang.org/protobuf/types/known/durationpb"
	_ "google.golang.org/protobuf/types/known/structpb"
	_ "google.golang.org/protobuf/types/known/wrapperspb"
)

func TestA(t *testing.T) {
	r, err := proto.MergedRegistry()

	fmt.Println(r.NumFiles())
	require.NoError(t, err)
}
