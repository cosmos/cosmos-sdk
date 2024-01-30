package prompt

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"cosmossdk.io/client/v2/autocli/flag"
	addresscodec "cosmossdk.io/core/address"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/manifoldco/promptui"
)

const GovModuleName = "gov"

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

		// create prompts
		prompt := promptui.Prompt{
			Label:    fmt.Sprintf("Enter %s %s", promptPrefix, fieldName),
			Validate: ValidatePromptNotEmpty,
		}

		// signer field
		if strings.EqualFold(fieldName, flag.GetSignerFieldName(msg.Descriptor())) {
			// pre-fill with gov address
			govAddr := address.Module(GovModuleName)
			govAddrStr, err := addressCodec.BytesToString(govAddr)
			if err != nil {
				return msg, fmt.Errorf("failed to convert gov address to string: %w", err)
			}

			// note, we don't set prompt.Validate here because we need to get the scalar annotation
			prompt.Default = govAddrStr
		}

		// validate address fields
		scalarField, ok := flag.GetScalarType(field)
		if ok {
			switch scalarField {
			case flag.AddressStringScalarType:
				prompt.Validate = func(input string) error {
					if _, err := addressCodec.StringToBytes(input); err != nil {
						return fmt.Errorf("invalid address")
					}

					return nil
				}
			case flag.ValidatorAddressStringScalarType:
				prompt.Validate = func(input string) error {
					if _, err := validatorAddressCodec.StringToBytes(input); err != nil {
						return fmt.Errorf("invalid validator address")
					}

					return nil
				}
			case flag.ConsensusAddressStringScalarType:
				prompt.Validate = func(input string) error {
					if _, err := consensusAddressCodec.StringToBytes(input); err != nil {
						return fmt.Errorf("invalid consensus address")
					}

					return nil
				}
			case flag.CoinScalarType:
				prompt.Validate = ValidatePromptCoins
			default:
				// prompt.Validate = ValidatePromptNotEmpty (we possibly don't want to force all fields to be non-empty)
				prompt.Validate = nil
			}
		}

		result, err := prompt.Run()
		if err != nil {
			return msg, fmt.Errorf("failed to prompt for %s: %w", fieldName, err)
		}

		switch field.Kind() {
		case protoreflect.StringKind:
			msg.Set(field, protoreflect.ValueOfString(result))
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			resultUint, err := strconv.ParseUint(result, 10, 0)
			if err != nil {
				return msg, fmt.Errorf("invalid value for int: %w", err)
			}

			msg.Set(field, protoreflect.ValueOfUint64(resultUint))
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind, protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			resultInt, err := strconv.ParseInt(result, 10, 0)
			if err != nil {
				return msg, fmt.Errorf("invalid value for int: %w", err)
			}
			// If a value was successfully parsed the ranges of:
			//      [minInt,     maxInt]
			// are within the ranges of:
			//      [minInt64, maxInt64]
			// of which on 64-bit machines, which are most common,
			// int==int64
			msg.Set(field, protoreflect.ValueOfInt64(resultInt))
		case protoreflect.BoolKind:
			resultBool, err := strconv.ParseBool(result)
			if err != nil {
				return msg, fmt.Errorf("invalid value for bool: %w", err)
			}

			msg.Set(field, protoreflect.ValueOfBool(resultBool))
		case protoreflect.MessageKind:
			// TODO
		default:
			// skip any other types
			continue // TODO(@julienrbrt) add support for other types
		}
	}

	return msg, nil
}

func PromptStruct[T any](promptPrefix string, data T) (T, error) {
	v := reflect.ValueOf(&data).Elem()
	if v.Kind() == reflect.Interface {
		v = reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
	}

	for i := 0; i < v.NumField(); i++ {
		// if the field is a struct skip or not slice of string or int then skip
		switch v.Field(i).Kind() {
		case reflect.Struct:
			// TODO(@julienrbrt) in the future we can add a recursive call to Prompt
			continue
		case reflect.Slice:
			if v.Field(i).Type().Elem().Kind() != reflect.String && v.Field(i).Type().Elem().Kind() != reflect.Int {
				continue
			}
		}

		// create prompts
		prompt := promptui.Prompt{
			Label:    fmt.Sprintf("Enter %s %s", promptPrefix, strings.Title(v.Type().Field(i).Name)), // nolint:staticcheck // strings.Title has a better API
			Validate: ValidatePromptNotEmpty,
		}

		fieldName := strings.ToLower(v.Type().Field(i).Name)

		result, err := prompt.Run()
		if err != nil {
			return data, fmt.Errorf("failed to prompt for %s: %w", fieldName, err)
		}

		switch v.Field(i).Kind() {
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
		default:
			// skip any other types
			continue
		}
	}

	return data, nil
}
