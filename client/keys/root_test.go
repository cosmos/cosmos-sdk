package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommands(t *testing.T) {
	rootCommands := Commands("home")
	assert.NotNil(t, rootCommands)

	// Commands are registered
	assert.Equal(t, 9, len(rootCommands.Commands()))
}
