package unstructured

import (
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/dynamicpb"
	"testing"
)

var pb *dynamicpb.Message

func BenchmarkMap_Marshal_Map(b *testing.B) {
	fdBuilder := protoimpl.DescBuilder{
		RawDescriptor: fileDescriptor,
		TypeResolver:  new(protoregistry.Types),
		FileRegistry:  new(protoregistry.Files),
	}

	md := fdBuilder.Build().File.Messages().ByName("WithMap")

	msg := Map{
		"a_map": map[int64]string{
			1: "hi",
		},
	}
	var err error

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = msg.Marshal(md)
		if err != nil {
			b.Fatal(err)
		}
	}

}
