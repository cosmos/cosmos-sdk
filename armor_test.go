package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleArmor(t *testing.T) {
	blockType := "MINT TEST"
	data := []byte("somedata")
	armorStr := EncodeArmor(blockType, nil, data)

	// Decode armorStr and test for equivalence.
	blockType2, _, data2, err := DecodeArmor(armorStr)
	require.Nil(t, err, "%+v", err)
	assert.Equal(t, blockType, blockType2)
	assert.Equal(t, data, data2)
}
