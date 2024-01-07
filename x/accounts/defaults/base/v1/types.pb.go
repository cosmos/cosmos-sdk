package v1

import (
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/gogoproto/proto"
)

type MsgInit struct {
	proto.Message
	PubKey []byte
}

type MsgInitResponse struct {
	proto.Message
}

type MsgSwapPubKey struct {
	proto.Message
	NewPubKey []byte
}

type MsgSwapPubKeyResponse struct {
	proto.Message
}

type AuthenticationData struct {
	proto.Message
	SignMode  signing.SignMode
	Signature []byte
}
