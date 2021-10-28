package kv

import "context"

type jointIterator struct {
	iters []Iterator
}

func (j jointIterator) Key() []byte {
	panic("implement me")
}

func (j jointIterator) Next() {
	panic("implement me")
}

func (j jointIterator) Valid() bool {
	panic("implement me")
}

func (j jointIterator) Close() {
	panic("implement me")
}

func (j jointIterator) Context() context.Context {
	return j.iters[0].Context()
}

// NewJointIterator returns an Iterator
func NewJointIterator(iters ...Iterator) Iterator {
	return jointIterator{iters: iters}
}
