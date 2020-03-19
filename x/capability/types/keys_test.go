package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/capability/types"
)

func TestRevCapabilityKey(t *testing.T) {
	expected := []byte("bank/rev/send")
	require.Equal(t, expected, types.RevCapabilityKey("bank", "send"))
}

func TestFwdCapabilityKey(t *testing.T) {
	cap := types.NewCapabilityKey(23)
	expected := []byte(fmt.Sprintf("bank/fwd/%s", cap))
	require.Equal(t, expected, types.FwdCapabilityKey("bank", cap))
}

func TestIndexToKey(t *testing.T) {
	require.Equal(t, []byte{0x5a, 0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, types.IndexToKey(3162))
}

func TestIndexFromKey(t *testing.T) {
	require.Equal(t, uint64(3162), types.IndexFromKey([]byte{0x5a, 0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}))
}
