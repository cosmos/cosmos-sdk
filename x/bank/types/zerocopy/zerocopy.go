package zerocopy

import "encoding/binary"

type Serializable[T any] interface {
	WithBufferContext(*BufferContext) *T
	BufferContext() *BufferContext
	Size() uint32
}

type Buffer struct {
	buf []byte
}

type BufferContext struct {
	parent *Buffer
	offset uint32
	buf    []byte
}

const DefaultBufferCapacity = 1024

func (b *Buffer) Alloc(n uint32) (offset uint32, ctx *BufferContext) {
	if b.buf == nil {
		b.buf = make([]byte, 0, DefaultBufferCapacity)
	}

	offset = uint32(len(b.buf))
	for i := uint32(0); i < n; i++ {
		b.buf = append(b.buf, 0)
	}
	ctx = &BufferContext{
		parent: b,
		offset: offset,
		buf:    b.buf[offset : offset+n],
	}
	return
}

func (b *Buffer) Resolve(offset uint32) *BufferContext {
	ctx := &BufferContext{
		parent: b,
		buf:    b.buf[offset:],
	}
	return ctx
}

func (c BufferContext) ReadUint32(offset uint32) uint32 {
	return binary.LittleEndian.Uint32(c.buf[offset : offset+8])
}

func (c BufferContext) SetUint32(offset uint32, value uint32) {
	binary.LittleEndian.PutUint32(c.buf[offset:offset+8], value)
}

func (c BufferContext) ReadUint64(offset uint32) uint64 {
	return binary.LittleEndian.Uint64(c.buf[offset : offset+8])
}

func (c BufferContext) SetUint64(offset uint32, value uint64) {
	binary.LittleEndian.PutUint64(c.buf[offset:offset+8], value)
}

func (c BufferContext) ResolvePointer(offset uint32) *BufferContext {
	ptrOffset := c.ReadUint32(offset)
	if ptrOffset == 0 {
		return nil
	}
	return c.parent.Resolve(ptrOffset)
}

func (c BufferContext) AllocPointer(offset uint32, size uint32) *BufferContext {
	ptrOffset, ctx := c.parent.Alloc(size)
	c.SetUint32(offset, ptrOffset)
	return ctx
}

func (c BufferContext) ReadString() string {
	n := c.ReadUint32(0)
	return string(c.buf[4 : 4+n])
}

func (c BufferContext) SetString(offset uint32, x string) {
	n := len(x)
	ptrOffset, ctx := c.parent.Alloc(uint32(n) + 4)
	c.SetUint32(offset, ptrOffset)
	binary.LittleEndian.PutUint32(ctx.buf[0:4], uint32(n))
	copy(ctx.buf[4:4+n], x)
}

func ReadArray[T any, PT interface {
	Serializable[T]
	*T
}](ctx *BufferContext, offset uint32) Array[*T] {
	arrayCtx := ctx.ResolvePointer(offset)
	if arrayCtx == nil {
		return Array[*T]{array: nil}
	}

	n := arrayCtx.ReadUint32(0)
	array := make([]*T, n)
	elemOffset := uint32(4)
	for i := uint32(0); i < n; i++ {
		elem := new(T)
		pt := PT(elem)
		size := pt.Size()
		parent := ctx.parent
		pt.WithBufferContext(&BufferContext{
			parent: parent,
			offset: elemOffset,
			buf:    parent.buf[elemOffset : elemOffset+size],
		})
		array[i] = elem
		elemOffset += size
	}

	return Array[*T]{array: array}
}

func InitArray[T any, PT interface {
	Serializable[T]
	*T
}](ctx *BufferContext, offset uint32, n int) Array[*T] {
	var elem PT
	arrayCtx := ctx.AllocPointer(offset, 4+uint32(n)*elem.Size())
	arrayCtx.SetUint32(0, uint32(n))
	array := make([]*T, n)
	elemOffset := uint32(4)
	for i := 0; i < n; i++ {
		elem := new(T)
		pt := PT(elem)
		size := pt.Size()
		parent := ctx.parent
		pt.WithBufferContext(&BufferContext{
			parent: parent,
			offset: elemOffset,
			buf:    parent.buf[elemOffset : elemOffset+size],
		})
		array[i] = elem
	}
	return Array[*T]{array: array}
}

type Array[T any] struct {
	array []T
}

func (a Array[T]) Len() int {
	return len(a.array)
}

func (a Array[T]) Get(i int) T {
	return a.array[i]
}
