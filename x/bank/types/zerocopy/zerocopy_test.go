package zerocopy

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestMsgSend(t *testing.T) {
	msgSend := &msgSend{}
	msgSend.SetFromAddress("cosmos1").SetToAddress("cosmos2")
	coins := msgSend.InitCoins(2)
	coins.Get(0).SetDenom("atom").SetAmount("100")
	coins.Get(1).SetDenom("foo").SetAmount("200")

	msgSend2 := &msgSend{}
	msgSend2.WithBufferContext(msgSend.BufferContext())
	assert.Equal(t, msgSend2.FromAddress(), "cosmos1")
	assert.Equal(t, msgSend2.ToAddress(), "cosmos2")
	coins2 := msgSend2.Coins()
	assert.Equal(t, coins2.Len(), 2)
	assert.Equal(t, coins2.Get(0).Denom(), "atom")
	assert.Equal(t, coins2.Get(0).Amount(), "100")
	assert.Equal(t, coins2.Get(1).Denom(), "foo")
	assert.Equal(t, coins2.Get(1).Amount(), "200")
}
