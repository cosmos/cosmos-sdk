package prompt

import (
	"cosmossdk.io/client/v2/autocli/flag"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"

	addresscodec "cosmossdk.io/core/address"
)

func Prompt(
	addressCodec addresscodec.Codec,
	validatorAddressCodec addresscodec.Codec,
	consensusAddressCodec addresscodec.Codec,
	promptPrefix string,
	msg protoreflect.Message,
) (protoreflect.Message, error) {
	fields := msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldName := string(field.Name())

		// create prompt with promptui

		// signer field
		if strings.EqualFold(fieldName, flag.GetSignerFieldName(msg.Descriptor())) {
			//	here signer must be set in some cases. For example gov module address but this prompt should work for any
			//	kind of message...
		}
	}
	return nil, nil
}
