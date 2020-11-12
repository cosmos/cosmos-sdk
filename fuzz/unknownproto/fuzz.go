package unknownproto

import (
	"github.com/cosmos/cosmos-sdk/codec/unknownproto"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func Fuzz(b []byte) int {
	msg := new(testdata.TestVersion2)
	resolver := new(unknownproto.DefaultAnyResolver)
	_, err1 := unknownproto.RejectUnknownFields(b, msg, true, resolver)
	_, err2 := unknownproto.RejectUnknownFields(b, msg, false, resolver)
	if (err1 != nil) != (err2 != nil) {
		return 1
	}
	return -1
}
