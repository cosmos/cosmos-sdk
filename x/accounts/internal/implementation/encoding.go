package implementation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cosmos/gogoproto/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

type Any = codectypes.Any

func FindMessageByName(name string) (ProtoMsg, error) {
	typ := proto.MessageType(name)
	if typ == nil {
		return nil, fmt.Errorf("no message type found for %s", name)
	}
	return reflect.New(typ.Elem()).Interface().(ProtoMsg), nil
}

func MessageName(msg ProtoMsg) string {
	return proto.MessageName(msg)
}

// PackAny packs a proto message into an anypb.Any.
func PackAny(msg ProtoMsg) (*Any, error) {
	return codectypes.NewAnyWithValue(msg)
}

// UnpackAny unpacks an anypb.Any into a proto message.
func UnpackAny[T any, PT ProtoMsgG[T]](anyPB *Any) (PT, error) {
	to := new(T)
	return to, UnpackAnyTo(anyPB, PT(to))
}

func UnpackAnyTo(anyPB *Any, to ProtoMsg) error {
	return proto.Unmarshal(anyPB.Value, to)
}

func UnpackAnyRaw(anyPB *Any) (proto.Message, error) {
	split := strings.Split(anyPB.TypeUrl, "/")
	name := split[len(split)-1]
	typ := proto.MessageType(name)
	if typ == nil {
		return nil, fmt.Errorf("no message type found for %s", name)
	}
	to := reflect.New(typ.Elem()).Interface().(proto.Message)
	return to, UnpackAnyTo(anyPB, to)
}

func Merge(a, b ProtoMsg) {
	proto.Merge(a, b)
}

func Equal(a, b ProtoMsg) bool {
	return proto.Equal(a, b)
}
