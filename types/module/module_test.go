package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetOrderBeginBlockers(t *testing.T) {
	mm := NewManager()
	mm.SetOrderBeginBlockers("a", "b", "c")
	obb := mm.OrderBeginBlockers
	require.Equal(t, 3, len(obb))
	assert.Equal(t, []string{"a", "b", "c"}, obb)
}
