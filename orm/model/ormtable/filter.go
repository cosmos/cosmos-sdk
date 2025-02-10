package ormtable

import "google.golang.org/protobuf/proto"

type filterIterator struct {
	Iterator
	filter func(proto.Message) bool
	msg    proto.Message
}

func (f *filterIterator) Next() bool {
	for f.Iterator.Next() {
		msg, err := f.Iterator.GetMessage()
		if err != nil {
			return false
		}

		if f.filter(msg) {
			f.msg = msg
			return true
		}
	}
	return false
}

func (f filterIterator) GetMessage() (proto.Message, error) {
	return f.msg, nil
}
