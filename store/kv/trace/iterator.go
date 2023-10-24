package trace

import (
	"io"

	"cosmossdk.io/store/v2"
)

var _ store.Iterator = (*iterator)(nil)

type iterator struct {
	parent  store.Iterator
	writer  io.Writer
	context store.TraceContext
}

func newIterator(w io.Writer, parent store.Iterator, tc store.TraceContext) store.Iterator {
	return &iterator{
		parent:  parent,
		writer:  w,
		context: tc,
	}
}

func (itr *iterator) Domain() ([]byte, []byte) {
	return itr.parent.Domain()
}

func (itr *iterator) Valid() bool {
	return itr.parent.Valid()
}

func (itr *iterator) Next() bool {
	return itr.parent.Next()
}

func (itr *iterator) Error() error {
	return itr.parent.Error()
}

func (itr *iterator) Close() {
	itr.parent.Close()
}

func (itr *iterator) Key() []byte {
	key := itr.parent.Key()

	writeOperation(itr.writer, IterKeyOp, itr.context, key, nil)
	return key
}

func (itr *iterator) Value() []byte {
	value := itr.parent.Value()

	writeOperation(itr.writer, IterValueOp, itr.context, nil, value)
	return value
}
