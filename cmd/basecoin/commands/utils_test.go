package commands

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHex(t *testing.T) {
	assert := assert.New(t)

	//test isHex
	hexNoPrefix := hex.EncodeToString([]byte("foobar"))
	hexWPrefix := "0x" + hexNoPrefix
	str := "foobar"
	strWPrefix := "0xfoobar"

	assert.True(isHex(hexWPrefix), "isHex not identifying hex with 0x prefix")
	assert.False(isHex(hexNoPrefix), "isHex shouldn't identify hex without 0x prefix")
	assert.False(isHex(str), "isHex shouldn't identify non-hex string")
	assert.False(isHex(strWPrefix), "isHex shouldn't identify non-hex string with 0x prefix")
	assert.True(StripHex(hexWPrefix) == hexNoPrefix, "StripHex doesn't remove first two characters")
}
