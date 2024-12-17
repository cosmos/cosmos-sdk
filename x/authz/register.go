package authz

import (
	"fmt"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GetProtoHTTPGetRuleMapping returns a mapping of proto method full name to it's HTTP GET annotation.
func GetProtoHTTPGetRuleMapping() map[string]string {
	protoFiles, err := gogoproto.MergedRegistry()
	if err != nil {
		panic(err)
	}

	httpGets := make(map[string]string)
	protoFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		for i := 0; i < fd.Services().Len(); i++ {
			// Get the service descriptor
			sd := fd.Services().Get(i)

			for j := 0; j < sd.Methods().Len(); j++ {
				// Get the method descriptor
				md := sd.Methods().Get(j)

				httpOption := proto.GetExtension(md.Options(), annotations.E_Http)
				if httpOption == nil {
					continue
				}

				httpRule, ok := httpOption.(*annotations.HttpRule)
				if !ok || httpRule == nil {
					continue
				}
				if httpRule.GetGet() == "" {
					continue
				}

				httpGets[httpRule.GetGet()] = string(md.FullName())
				fmt.Printf("service: %q \t get option: %q\n", md.FullName(), httpRule.GetGet())
			}
		}
		return true
	})

	return httpGets
}
