//go:build !race
// +build !race

// Disabled -race because the package github.com/manifoldco/promptui@v0.9.0
// has a data race and this code exposes it, but fixing it would require
// holding up the associated change to this.

package prompt_test

import (
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/chzyer/readline"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/client/v2/autocli/prompt"
	"cosmossdk.io/client/v2/internal/testpb"
)

// Tests that we successfully report overflows in parsing ints
// See https://github.com/cosmos/cosmos-sdk/issues/13346
func TestPromptIntegerOverflow(t *testing.T) {
	// Intentionally sending values out of the range of int.
	intOverflowers := []string{
		"-9223372036854775809",
		"9223372036854775808",
		"9923372036854775809",
		"-9923372036854775809",
		"18446744073709551616",
		"-18446744073709551616",
	}

	for _, intOverflower := range intOverflowers {
		overflowStr := intOverflower
		t.Run(overflowStr, func(t *testing.T) {
			origStdin := readline.Stdin
			defer func() {
				readline.Stdin = origStdin
			}()

			fin, fw := readline.NewFillableStdin(os.Stdin)
			readline.Stdin = fin
			_, err := fw.Write([]byte(overflowStr + "\n"))
			require.NoError(t, err)

			v, err := prompt.Prompt(mockAddressCodec{}, mockAddressCodec{}, mockAddressCodec{}, "", (&testpb.MsgRequest{}).ProtoReflect())
			require.Equal(t, (&testpb.MsgRequest{}).ProtoReflect(), v, "expected a value of zero")
			require.NotNil(t, err, "expected a report of an overflow")
			require.Contains(t, err.Error(), "range")
		})
	}
}

func TestPromptParseInteger(t *testing.T) {
	// Intentionally sending a value out of the range of
	values := []struct {
		in   string
		want int
	}{
		{fmt.Sprintf("%d", math.MinInt), math.MinInt},
		{"19991", 19991},
		{"991000000199", 991000000199},
	}

	for _, tc := range values {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			origStdin := readline.Stdin
			defer func() {
				readline.Stdin = origStdin
			}()

			fin, fw := readline.NewFillableStdin(os.Stdin)
			readline.Stdin = fin
			_, err := fw.Write([]byte(tc.in + "\n"))
			require.NoError(t, err)
			v, err := prompt.Prompt(mockAddressCodec{}, mockAddressCodec{}, mockAddressCodec{}, "", (&testpb.MsgRequest{}).ProtoReflect())
			require.Nil(t, err, "expected a nil error")
			require.NotNil(t, v)
			// require.Equal(t, tc.want, v.I, "expected %d = %d", tc.want, v.I)
		})
	}
}

type mockAddressCodec struct{}

func (mockAddressCodec) BytesToString([]byte) (string, error) {
	return "cosmos1y74p8wyy4enfhfn342njve6cjmj5c8dtl6emdk", nil
}

func (mockAddressCodec) StringToBytes(string) ([]byte, error) {
	return nil, nil
}
