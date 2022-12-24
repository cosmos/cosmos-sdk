package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommands(t *testing.T) {
	t.Parallel()
	rootCommands := Commands("home")
	assert.NotNil(t, rootCommands)

	// Commands are registered
	assert.Equal(t, 10, len(rootCommands.Commands()))
}
