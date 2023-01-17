package keys

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommands(t *testing.T) {
	rootCommands := Commands("home")
	assert.Assert(t, rootCommands != nil)

	// Commands are registered
	assert.Equal(t, 11, len(rootCommands.Commands()))
}
