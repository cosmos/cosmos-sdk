package crypto

import (
	"bytes"
	"testing"
)

func TestSimpleArmor(t *testing.T) {
	blockType := "MINT TEST"
	data := []byte("somedata")
	armorStr := EncodeArmor(blockType, nil, data)
	t.Log("Got armor: ", armorStr)

	// Decode armorStr and test for equivalence.
	blockType2, _, data2, err := DecodeArmor(armorStr)
	if err != nil {
		t.Error(err)
	}
	if blockType != blockType2 {
		t.Errorf("Expected block type %v but got %v", blockType, blockType2)
	}
	if !bytes.Equal(data, data2) {
		t.Errorf("Expected data %X but got %X", data2, data)
	}
}
