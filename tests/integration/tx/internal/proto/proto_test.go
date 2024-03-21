package proto

import (
	"cosmossdk.io/math"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/protoadapt"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"testing"
)

func Test_ProtoStuff(t *testing.T) {
	m := &types.Coin{
		Denom:  "denom",
		Amount: math.NewInt(100),
	}

	mv2 := protoadapt.MessageV2Of(m)
	desc := mv2.ProtoReflect().Descriptor()
	dm := dynamicpb.NewMessage(desc)
	for i := 0; i < desc.Fields().Len(); i++ {
		field := desc.Fields().Get(i)
		v := dm.ProtoReflect().Get(field)
		if field.Name() == "amount" {
			dm.Set(field, protoreflect.ValueOf("100"))
		}
		fmt.Println(field.Name(), v)
	}
	fmt.Println(dm)
}
