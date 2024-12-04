package prompt

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/client/v2/autocli/flag"
	"cosmossdk.io/client/v2/internal/prompt"
	addresscodec "cosmossdk.io/core/address"
)

// PromptMessage prompts the user for values to populate a protobuf message interactively.
// It returns the populated message and any error encountered during prompting.
func PromptMessage(
	addressCodec addresscodec.Codec, validatorAddressCodec addresscodec.Codec,
	consensusAddressCodec addresscodec.Codec, promptPrefix string, msg protoreflect.Message,
) (protoreflect.Message, error) {
	return promptMessage(addressCodec, validatorAddressCodec, consensusAddressCodec, promptPrefix, nil, msg)
}

// promptMessage prompts the user for values to populate a protobuf message interactively.
// stdIn is provided to make the function easier to unit test by allowing injection of predefined inputs.
func promptMessage(
	addressCodec addresscodec.Codec, validatorAddressCodec addresscodec.Codec,
	consensusAddressCodec addresscodec.Codec, promptPrefix string,
	stdIn io.ReadCloser, msg protoreflect.Message,
) (protoreflect.Message, error) {
	promptUi := promptui.Prompt{
		Validate: prompt.ValidatePromptNotEmpty,
		Stdin:    stdIn,
	}

	fields := msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldName := string(field.Name())

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
				promptUi.Validate = prompt.ValidateAddress(addressCodec)
			case flag.ValidatorAddressStringScalarType:
				promptUi.Validate = prompt.ValidateAddress(validatorAddressCodec)
			case flag.ConsensusAddressStringScalarType:
				promptUi.Validate = prompt.ValidateAddress(consensusAddressCodec)
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

// valueOf prompts the user for a comma-separated list of values for a repeated field.
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
	validatorAddressCodec addresscodec.Codec, consensusAddressCodec addresscodec.Codec,
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
	validatorAddressCodec addresscodec.Codec, consensusAddressCodec addresscodec.Codec,
	promptPrefix string, stdIn io.ReadCloser, msg protoreflect.Message,
) error {
	fieldName := promptPrefix + "." + string(f.Name())
	nestedMsg := msg.Get(f).Message()
	//if nestedMsg.IsValid() {
	//	nestedMsg = nestedMsg.New()
	//} else {
	//	nestedMsg = msg.Get(f).Message()
	//}
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
	validatorAddressCodec addresscodec.Codec, consensusAddressCodec addresscodec.Codec,
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

// PromptStruct prompts for values of a struct's fields interactively.
// It returns the populated struct and any error encountered.
func PromptStruct[T any](promptPrefix string, data T) (T, error) {
	return promptStruct(promptPrefix, data, nil)
}

// promptStruct prompts for values of a struct's fields interactively.
//
// For each field in the struct:
// - Pointer fields are initialized if nil and handled recursively if they contain structs
// - Struct fields are handled recursively
// - String and int slices are supported
// - String and int fields are prompted for and populated
// - Only String and int pointers are supported
// - Other types are skipped
func promptStruct[T any](promptPrefix string, data T, stdIn io.ReadCloser) (T, error) {
	v := reflect.ValueOf(&data).Elem()
	if v.Kind() == reflect.Interface {
		v = reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := strings.ToLower(v.Type().Field(i).Name)

		// Handle pointer types
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			if field.Elem().Kind() == reflect.Struct {
				result, err := promptStruct(promptPrefix+"."+fieldName, field.Interface(), stdIn)
				if err != nil {
					return data, err
				}
				field.Set(reflect.ValueOf(result))
				continue
			}
		}

		switch field.Kind() {
		case reflect.Struct:
			// For struct fields, create a new pointer to handle them
			structPtr := reflect.New(field.Type()).Interface()
			reflect.ValueOf(structPtr).Elem().Set(field)

			result, err := promptStruct(promptPrefix+"."+fieldName, structPtr, stdIn)
			if err != nil {
				return data, err
			}

			// Get the actual struct value from the result
			resultValue := reflect.ValueOf(result)
			if resultValue.Kind() == reflect.Ptr {
				resultValue = resultValue.Elem()
			}
			field.Set(resultValue)
			continue
		case reflect.Slice:
			if v.Field(i).Type().Elem().Kind() != reflect.String && v.Field(i).Type().Elem().Kind() != reflect.Int {
				continue
			}
		}

		// create prompts
		prompt := promptui.Prompt{
			Label:    fmt.Sprintf("Enter %s %s", promptPrefix, strings.Title(fieldName)), // nolint:staticcheck // strings.Title has a better API
			Validate: prompt.ValidatePromptNotEmpty,
			Stdin:    stdIn,
		}

		result, err := prompt.Run()
		if err != nil {
			return data, fmt.Errorf("failed to prompt for %s: %w", fieldName, err)
		}

		switch field.Kind() {
		case reflect.String:
			v.Field(i).SetString(result)
		case reflect.Int:
			resultInt, err := strconv.ParseInt(result, 10, 0)
			if err != nil {
				return data, fmt.Errorf("invalid value for int: %w", err)
			}
			v.Field(i).SetInt(resultInt)
		case reflect.Slice:
			switch v.Field(i).Type().Elem().Kind() {
			case reflect.String:
				v.Field(i).Set(reflect.ValueOf([]string{result}))
			case reflect.Int:
				resultInt, err := strconv.ParseInt(result, 10, 0)
				if err != nil {
					return data, fmt.Errorf("invalid value for int: %w", err)
				}

				v.Field(i).Set(reflect.ValueOf([]int{int(resultInt)}))
			}
		case reflect.Ptr:
			// Handle pointer fields by creating a new value and setting it
			ptrValue := reflect.New(field.Type().Elem())
			if ptrValue.Elem().Kind() == reflect.String {
				ptrValue.Elem().SetString(result)
				v.Field(i).Set(ptrValue)
			} else if ptrValue.Elem().Kind() == reflect.Int {
				resultInt, err := strconv.ParseInt(result, 10, 0)
				if err != nil {
					return data, fmt.Errorf("invalid value for int: %w", err)
				}
				ptrValue.Elem().SetInt(resultInt)
				v.Field(i).Set(ptrValue)
			}
		default:
			// skip any other types
			continue
		}
	}

	return data, nil
}
