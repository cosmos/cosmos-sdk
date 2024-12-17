package grpcgateway

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThing(t *testing.T) {
	bz, err := os.ReadFile("mapping.json")
	require.NoError(t, err)
	var mapping map[string]string
	err = json.Unmarshal(bz, &mapping)
	require.NoError(t, err)

	match := matchURI("/cosmos/bank/v1beta1/denom_owners/ibc/denom/1", mapping)
	fmt.Println(match)
}
