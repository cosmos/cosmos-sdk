package internal

import (
	"google.golang.org/protobuf/proto"

	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/tests/integration/tx/internal/pulsar/testpb"
)

func ProvideCustomGetSigner() signing.CustomGetSigner {
	return signing.CustomGetSigner{
		MsgType: proto.MessageName(&testpb.TestRepeatedFields{}),
		Fn: func(msg proto.Message) ([][]byte, error) {
			testMsg := msg.(*testpb.TestRepeatedFields)
			// arbitrary logic
			signer := testMsg.NullableDontOmitempty[1].Value
			return [][]byte{[]byte(signer)}, nil
		},
	}
}
