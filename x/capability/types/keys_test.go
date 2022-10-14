package types_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/capability/types"
)

func TestRevCapabilityKey(t *testing.T) {
	expected := []byte("bank/rev/send")
	require.Equal(t, expected, types.RevCapabilityKey("bank", "send"))
}

func TestFwdCapabilityKey(t *testing.T) {
	cap := types.NewCapability(23)
	key := fmt.Sprintf("%#010p", cap)
	if len(key) > 10 {
		key = key[len(key)-10:]
	}
	require.Equal(t, 10, len(key))
	expected := []byte(fmt.Sprintf("bank/fwd/0x%s", key))
	require.Equal(t, expected, types.FwdCapabilityKey("bank", cap))
}

func TestIndexToKey(t *testing.T) {
	require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xc, 0x5a}, types.IndexToKey(3162))
}

func TestIndexFromKey(t *testing.T) {
	require.Equal(t, uint64(3162), types.IndexFromKey([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xc, 0x5a}))
}

// to test the backward compatibiltiy of the new function
func legacyFwdCapabilityKey(module string, cap *types.Capability) []byte {
	return []byte(fmt.Sprintf("%s/fwd/%p", module, cap))
}

func TestFwdCapabilityKeyCompatibility(t *testing.T) {
	cap := types.NewCapability(24)
	new := types.FwdCapabilityKey("bank", cap)
	old := legacyFwdCapabilityKey("bank", cap)
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm" {
		// the legacy version has 1 more byte on mac m1
		require.Equal(t, len(old), len(new)+1)
	} else {
		// otherwise, the new version is identical
		require.Equal(t, new, old)
	}
}
