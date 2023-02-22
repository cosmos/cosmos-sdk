package zerocopy

const (
	coinDenomOffset  = 0
	coinAmountOffset = coinDenomOffset + 4
	coinSize         = coinAmountOffset + 4
)

type Coin struct {
	ctx *BufferContext
}

func (c *Coin) Denom() string {
	return c.ctx.ResolvePointer(coinDenomOffset).ReadString()
}

func (c *Coin) Amount() string {
	return c.ctx.ResolvePointer(coinAmountOffset).ReadString()
}

func (c *Coin) SetDenom(x string) *Coin {
	c.ctx.SetString(coinDenomOffset, x)
	return c
}

func (c *Coin) SetAmount(x string) *Coin {
	c.ctx.SetString(coinAmountOffset, x)
	return c
}

func (c *Coin) WithBufferContext(ctx *BufferContext) *Coin {
	c.ctx = ctx
	return c
}

func (c *Coin) BufferContext() *BufferContext {
	return c.ctx
}

func (c *Coin) Size() uint32 {
	return coinSize
}
