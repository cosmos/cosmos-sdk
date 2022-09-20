package cli_test

import (
	"sync"
	"testing"

	"github.com/chzyer/readline"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
)

type st struct {
	ToOverflow int
}

// On the tests running in Github actions, somehow there are races.
var globalMu sync.Mutex

// Tests that we successfully report overflows in parsing ints
// See https://github.com/cosmos/cosmos-sdk/issues/13346
func TestPromptIntegerOverflow(t *testing.T) {
	// Intentionally sending a value out of the range of
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
			readline.Stdout.Write([]byte(overflowStr + "\n"))

			v, err := cli.Prompt(st{}, "")
			assert.NotNil(t, err, "expected a report of an overflow")
			assert.Equal(t, st{}, v, "expected a value of zero")
		})
	}
}
