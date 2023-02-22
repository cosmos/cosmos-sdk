package zerocopy

const (
	msgSendFromAddressOffset = 0
	msgSendToAddressOffset   = msgSendFromAddressOffset + 4
	msgSendCoinsOffset       = msgSendToAddressOffset + 4
	msgSendSize              = msgSendCoinsOffset + 4
)

type MsgSend struct {
	ctx *BufferContext
}

func (m *MsgSend) FromAddress() string {
	if m.ctx == nil {
		return ""
	}
	return m.ctx.ResolvePointer(msgSendFromAddressOffset).ReadString()
}

func (m *MsgSend) ToAddress() string {
	if m.ctx == nil {
		return ""
	}
	return m.ctx.ResolvePointer(msgSendToAddressOffset).ReadString()
}

func (m *MsgSend) init() {
	if m.ctx == nil {
		_, m.ctx = (&Buffer{}).Alloc(msgSendSize)
	}
}

func (m *MsgSend) SetFromAddress(x string) *MsgSend {
	m.init()
	m.ctx.SetString(msgSendFromAddressOffset, x)
	return m
}

func (m *MsgSend) SetToAddress(x string) *MsgSend {
	m.init()
	m.ctx.SetString(msgSendToAddressOffset, x)
	return m
}

func (m *MsgSend) Coins() Array[*Coin] {
	m.init()
	return ReadArray[Coin](m.ctx, msgSendCoinsOffset)
}

func (m *MsgSend) InitCoins(size int) Array[*Coin] {
	m.init()
	return InitArray[Coin](m.ctx, msgSendCoinsOffset, size)
}

func (m *MsgSend) WithContext(ctx *BufferContext) *MsgSend {
	m.ctx = ctx
	return m
}

func (c *MsgSend) WithBufferContext(ctx *BufferContext) *MsgSend {
	c.ctx = ctx
	return c
}

func (c *MsgSend) BufferContext() *BufferContext {
	return c.ctx
}

func (c *MsgSend) Size() uint32 {
	return msgSendSize
}
