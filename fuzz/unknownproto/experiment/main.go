package main

import (
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/codec/unknownproto"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func main() {
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	msg := new(testdata.TestVersion2)
	resolver := new(unknownproto.DefaultAnyResolver)
	_, err1 := unknownproto.RejectUnknownFields(b, msg, true, resolver)
	_, err2 := unknownproto.RejectUnknownFields(b, msg, false, resolver)
	panic(err1)
	panic(err2)
}
