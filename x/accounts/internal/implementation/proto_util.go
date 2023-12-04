package implementation

import (
	"github.com/cosmos/cosmos-proto/anyutil"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// PackAny packs a proto message into an anypb.Any.
func PackAny(msg proto.Message) (*anypb.Any, error) {
	anyPB := new(anypb.Any)
	return anyPB, anyutil.MarshalFrom(anyPB, msg, proto.MarshalOptions{Deterministic: true})
}

// UnpackAny unpacks an anypb.Any into a proto message.
func UnpackAny[T any, PT ProtoMsg[T]](anyPB *anypb.Any) (PT, error) {
	to := new(T)
	return to, UnpackAnyTo(anyPB, PT(to))
}

func UnpackAnyTo(anyPB *anypb.Any, to proto.Message) error {
	return anypb.UnmarshalTo(anyPB, to, proto.UnmarshalOptions{
		DiscardUnknown: true,
	})
}

func UnpackAnyRaw(anyPB *anypb.Any) (proto.Message, error) {
	return anypb.UnmarshalNew(anyPB, proto.UnmarshalOptions{
		DiscardUnknown: true,
	})
}
