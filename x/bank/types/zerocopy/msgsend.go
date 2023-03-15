package zerocopy

const (
	msgSendFromAddressOffset = 0
	msgSendToAddressOffset   = msgSendFromAddressOffset + 4
	msgSendCoinsOffset       = msgSendToAddressOffset + 4
	msgSendSize              = msgSendCoinsOffset + 4
)

type msgSend struct {
	ctx *BufferContext
}

type MsgSend interface {
	FromAddress() (string, error)
	SetFromAddress(string) MsgSend
	ToAddress() (string, error)
	SetToAddress(string) MsgSend
	Coins() (Array[*Coin], error)
}

func (m *msgSend) FromAddress() string {
	if m.ctx == nil {
		return ""
	}
	return m.ctx.ResolvePointer(msgSendFromAddressOffset).ReadString()
}

func (m *msgSend) ToAddress() string {
	if m.ctx == nil {
		return ""
	}
	return m.ctx.ResolvePointer(msgSendToAddressOffset).ReadString()
}

func (m *msgSend) init() {
	if m.ctx == nil {
		_, m.ctx = (&Buffer{}).Alloc(msgSendSize)
	}
}

func (m *msgSend) SetFromAddress(x string) *msgSend {
	m.init()
	m.ctx.SetString(msgSendFromAddressOffset, x)
	return m
}

func (m *msgSend) SetToAddress(x string) *msgSend {
	m.init()
	m.ctx.SetString(msgSendToAddressOffset, x)
	return m
}

func (m *msgSend) Coins() Array[*Coin] {
	m.init()
	return ReadArray[Coin](m.ctx, msgSendCoinsOffset)
}

func (m *msgSend) InitCoins(size int) Array[*Coin] {
	m.init()
	return InitArray[Coin](m.ctx, msgSendCoinsOffset, size)
}

func (m *msgSend) WithContext(ctx *BufferContext) *msgSend {
	m.ctx = ctx
	return m
}

func (m *msgSend) WithBufferContext(ctx *BufferContext) *msgSend {
	m.ctx = ctx
	return m
}

func (m *msgSend) BufferContext() *BufferContext {
	return m.ctx
}

func (m *msgSend) Size() uint32 {
	return msgSendSize
}
