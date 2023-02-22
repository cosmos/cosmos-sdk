package zerocopy

import "encoding/binary"

type Context struct {
	buf []byte
}

func (c Context) ReadUint32(offset uint64) uint32 {
	return binary.LittleEndian.Uint32(c.buf[offset : offset+8])
}

func (c Context) SetUint32(offset uint64, value uint32) {
	binary.LittleEndian.PutUint32(c.buf[offset:offset+8], value)
}

func (c Context) ReadUint64(offset uint64) uint64 {
	return binary.LittleEndian.Uint64(c.buf[offset : offset+8])
}

func (c Context) SetUint64(offset uint64, value uint64) {
	binary.LittleEndian.PutUint64(c.buf[offset:offset+8], value)
}

func (c Context) ResolvePointer(offset uint64) Context {
	ptrOffset := c.ReadUint64(offset)
	return Context{c.buf[ptrOffset:]}
}

func (c Context) ReadString() string {
	n := c.ReadUint32(0)
	return string(c.buf[4 : 4+n])
}

func (c Context) SetString(x string) {
	n := len(x)
	binary.LittleEndian.AppendUint32(c.buf, uint32(n))
	// TODO deal with making sure there's enough space in array
	copy(c.buf[4:4+n], x)
}

type MsgSend struct {
	ctx Context
}

func (m MsgSend) FromAddress() string {
	return m.ctx.ResolvePointer(0).ReadString()
}

func (m MsgSend) ToAddress() string {
	return m.ctx.ResolvePointer(4).ReadString()
}

func (m MsgSend) SetFromAddress(x string) MsgSend {
	m.ctx.ResolvePointer(0).SetString(x)
	return m
}

func (m MsgSend) SetToAddress(x string) MsgSend {
	m.ctx.ResolvePointer(4).SetString(x)
	return m
}

func (m MsgSend) Coins() Array[Coin] {
	return Array[Coin]{m.ctx.ResolvePointer(8)}
}

func (m MsgSend) WithContext(ctx Context) MsgSend {
	m.ctx = ctx
	return m
}

type Coin struct {
	ctx Context
}

func (c Coin) Denom() string {
	return c.ctx.ResolvePointer(0).ReadString()
}

func (c Coin) Amount() string {
	return c.ctx.ResolvePointer(4).ReadString()
}

func (c Coin) SetDenom(x string) Coin {
	c.ctx.ResolvePointer(0).SetString(x)
	return c
}

func (c Coin) SetAmount(x string) Coin {
	c.ctx.ResolvePointer(4).SetString(x)
	return c
}

func (c Coin) WithContext(ctx Context) Coin {
	c.ctx = ctx
	return c
}

type Array[T WithContext[T]] struct {
	ctx Context
}

func Get[T WithContext[T]](array Array[T], i int) T {
	var x T
	array.ctx.ResolvePointer(uint64((i + 1) * 4))
	return x.WithContext(array.ctx)
}

func Length[T any](array Array[T]) int {
	return int(array.ctx.ReadUint64(0))
}

func Append[T any](array Array[T]) T {
	panic("TODO")
}

type WithContext[T any] interface {
	WithContext(Context) T
}
