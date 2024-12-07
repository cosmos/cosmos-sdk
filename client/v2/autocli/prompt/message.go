package prompt

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/client/v2/autocli/flag"
	addresscodec "cosmossdk.io/core/address"
)

// PromptMessage prompts the user for values to populate a protobuf message interactively.
// It returns the populated message and any error encountered during prompting.
func PromptMessage(
	addressCodec, validatorAddressCodec, consensusAddressCodec addresscodec.Codec,
	promptPrefix string, msg protoreflect.Message,
) (protoreflect.Message, error) {
	return promptMessage(addressCodec, validatorAddressCodec, consensusAddressCodec, promptPrefix, nil, msg)
}

// promptMessage prompts the user for values to populate a protobuf message interactively.
// stdIn is provided to make the function easier to unit test by allowing injection of predefined inputs.
func promptMessage(
	addressCodec, validatorAddressCodec, consensusAddressCodec addresscodec.Codec,
	promptPrefix string, stdIn io.ReadCloser, msg protoreflect.Message,
) (protoreflect.Message, error) {
	fields := msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldName := string(field.Name())

		promptUi := promptui.Prompt{
			Validate: ValidatePromptNotEmpty,
			Stdin:    stdIn,
		}

		// If this signer field has already a valid default value set,
		// use that value as the default prompt value. This is useful for
		// commands that have an authority such as gov.
		if strings.EqualFold(fieldName, flag.GetSignerFieldName(msg.Descriptor())) {
			if defaultValue := msg.Get(field); defaultValue.IsValid() {
				promptUi.Default = defaultValue.String()
			}
		}

		// validate address fields
		scalarField, ok := flag.GetScalarType(field)
		if ok {
			switch scalarField {
			case flag.AddressStringScalarType:
				promptUi.Validate = ValidateAddress(addressCodec)
			case flag.ValidatorAddressStringScalarType:
				promptUi.Validate = ValidateAddress(validatorAddressCodec)
			case flag.ConsensusAddressStringScalarType:
				promptUi.Validate = ValidateAddress(consensusAddressCodec)
			default:
				// prompt.Validate = ValidatePromptNotEmpty (we possibly don't want to force all fields to be non-empty)
				promptUi.Validate = nil
			}
		}

		// handle nested message fields recursively
		if field.Kind() == protoreflect.MessageKind {
			err := promptInnerMessageKind(field, addressCodec, validatorAddressCodec, consensusAddressCodec, promptPrefix, stdIn, msg)
			if err != nil {
				return nil, err
			}
			continue
		}

		// handle repeated fields by prompting for a comma-separated list of values
		if field.IsList() {
			list, err := promptList(field, msg, promptUi, promptPrefix)
			if err != nil {
				return nil, err
			}

			msg.Set(field, protoreflect.ValueOfList(list))
			continue
		}

		promptUi.Label = fmt.Sprintf("Enter %s %s", promptPrefix, fieldName)
		result, err := promptUi.Run()
		if err != nil {
			return msg, fmt.Errorf("failed to prompt for %s: %w", fieldName, err)
		}

		v, err := valueOf(field, result)
		if err != nil {
			return msg, err
		}
		msg.Set(field, v)
	}

	return msg, nil
}

// valueOf converts a string input value to a protoreflect.Value based on the field's type.
// It handles string, numeric, bool, bytes and enum field types.
// Returns the converted value and any error that occurred during conversion.
func valueOf(field protoreflect.FieldDescriptor, result string) (protoreflect.Value, error) {
	switch field.Kind() {
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(result), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		resultUint, err := strconv.ParseUint(result, 10, 0)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("invalid value for int: %w", err)
		}

		return protoreflect.ValueOfUint64(resultUint), nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind, protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		resultInt, err := strconv.ParseInt(result, 10, 0)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("invalid value for int: %w", err)
		}
		// If a value was successfully parsed the ranges of:
		//      [minInt,     maxInt]
		// are within the ranges of:
		//      [minInt64, maxInt64]
		// of which on 64-bit machines, which are most common,
		// int==int64
		return protoreflect.ValueOfInt64(resultInt), nil
	case protoreflect.BoolKind:
		resultBool, err := strconv.ParseBool(result)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("invalid value for bool: %w", err)
		}

		return protoreflect.ValueOfBool(resultBool), nil
	case protoreflect.BytesKind:
		resultBytes := []byte(result)
		return protoreflect.ValueOfBytes(resultBytes), nil
	case protoreflect.EnumKind:
		enumValue := field.Enum().Values().ByName(protoreflect.Name(result))
		if enumValue == nil {
			return protoreflect.Value{}, fmt.Errorf("invalid enum value %q", result)
		}
		return protoreflect.ValueOfEnum(enumValue.Number()), nil
	default:
		// TODO: add more kinds
		// skip any other types
		return protoreflect.Value{}, nil
	}
}

// promptList prompts the user for a comma-separated list of values for a repeated field.
// The user will be prompted to enter values separated by commas which will be parsed
// according to the field's type using valueOf.
func promptList(field protoreflect.FieldDescriptor, msg protoreflect.Message, promptUi promptui.Prompt, promptPrefix string) (protoreflect.List, error) {
	promptUi.Label = fmt.Sprintf("Enter %s %s list (separate values with ',')", promptPrefix, string(field.Name()))
	result, err := promptUi.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to prompt for %s: %w", string(field.Name()), err)
	}

	list := msg.Mutable(field).List()
	for _, item := range strings.Split(result, ",") {
		v, err := valueOf(field, item)
		if err != nil {
			return nil, err
		}
		list.Append(v)
	}

	return list, nil
}

// promptInnerMessageKind handles prompting for fields that are of message kind.
// It handles both single messages and repeated message fields by delegating to
// promptInnerMessage and promptMessageList respectively.
func promptInnerMessageKind(
	f protoreflect.FieldDescriptor, addressCodec addresscodec.Codec,
	validatorAddressCodec, consensusAddressCodec addresscodec.Codec,
	promptPrefix string, stdIn io.ReadCloser, msg protoreflect.Message,
) error {
	if f.IsList() {
		return promptMessageList(f, addressCodec, validatorAddressCodec, consensusAddressCodec, promptPrefix, stdIn, msg)
	}
	return promptInnerMessage(f, addressCodec, validatorAddressCodec, consensusAddressCodec, promptPrefix, stdIn, msg)
}

// promptInnerMessage prompts for a single nested message field. It creates a new message instance,
// recursively prompts for its fields, and sets the populated message on the parent message.
func promptInnerMessage(
	f protoreflect.FieldDescriptor, addressCodec addresscodec.Codec,
	validatorAddressCodec, consensusAddressCodec addresscodec.Codec,
	promptPrefix string, stdIn io.ReadCloser, msg protoreflect.Message,
) error {
	fieldName := promptPrefix + "." + string(f.Name())
	nestedMsg := msg.Get(f).Message()
	nestedMsg = nestedMsg.New()
	// Recursively prompt for nested message fields
	updatedMsg, err := promptMessage(
		addressCodec,
		validatorAddressCodec,
		consensusAddressCodec,
		fieldName,
		stdIn,
		nestedMsg,
	)
	if err != nil {
		return fmt.Errorf("failed to prompt for nested message %s: %w", fieldName, err)
	}

	msg.Set(f, protoreflect.ValueOfMessage(updatedMsg))
	return nil
}

// promptMessageList prompts for a repeated message field by repeatedly creating new message instances,
// prompting for their fields, and appending them to the list until the user chooses to stop.
func promptMessageList(
	f protoreflect.FieldDescriptor, addressCodec addresscodec.Codec,
	validatorAddressCodec, consensusAddressCodec addresscodec.Codec,
	promptPrefix string, stdIn io.ReadCloser, msg protoreflect.Message,
) error {
	list := msg.Mutable(f).List()
	for {
		fieldName := promptPrefix + "." + string(f.Name())
		// Create and populate a new message for the list
		nestedMsg := list.NewElement().Message()
		updatedMsg, err := promptMessage(
			addressCodec,
			validatorAddressCodec,
			consensusAddressCodec,
			fieldName,
			stdIn,
			nestedMsg,
		)
		if err != nil {
			return fmt.Errorf("failed to prompt for list item in %s: %w", fieldName, err)
		}

		list.Append(protoreflect.ValueOfMessage(updatedMsg))

		// Prompt whether to continue
		// TODO: may be better yes/no rather than interactive?
		continuePrompt := promptui.Select{
			Label: "Add another item?",
			Items: []string{"No", "Yes"},
			Stdin: stdIn,
		}

		_, result, err := continuePrompt.Run()
		if err != nil {
			return fmt.Errorf("failed to prompt for continuation: %w", err)
		}

		if result == "No" {
			break
		}
	}

	return nil
}
