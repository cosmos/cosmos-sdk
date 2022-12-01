package amino

import (
	fmt "fmt"
	"reflect"
	"strings"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"cosmossdk.io/api/amino"
	"github.com/cosmos/cosmos-sdk/codec"
)

// ValidateAminoAnnotations validates the `amino.*` protobuf annotations. It
// performs the following validations:
//   - Make sure `amino.name` is equal to the name in the Amino codec's registry.
//
// If `fileResolver` is nil, then protoregistry.GlobalFile will be used.
func ValidateAminoAnnotations(fdFiles *protoregistry.Files, aminoCdc *codec.LegacyAmino) error {
	var err error

	// Range through all files, and for each file, range through all its
	// messages to check the amino annotation.
	fdFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		for i := 0; i < fd.Messages().Len(); i++ {
			md := fd.Messages().Get(i)
			aminoName, found := proto.GetExtension(md.Options(), amino.E_Name).(string)
			if !found || aminoName == "" {
				continue
			}

			gogoMsgType := gogoproto.MessageType(string(md.FullName()))
			gogoMsg := reflect.New(gogoMsgType.Elem()).Interface()

			jsonBz, innerErr := aminoCdc.MarshalJSON(gogoMsg)
			if innerErr != nil {
				err = innerErr
				return false
			}

			// If the Amino JSON output has the "type" field, it means that the
			// underlying type has been registered in amino's registry.
			if !strings.HasPrefix(string(jsonBz), fmt.Sprintf(`{"type":"%s",`, aminoName)) {
				err = fmt.Errorf("proto message %s has incorrectly registered amino name %s", aminoName, md.FullName())
			}
		}

		return true
	})

	return err
}
