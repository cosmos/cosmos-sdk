package signing

import (
	"errors"
	"fmt"

	cosmos_proto "github.com/cosmos/cosmos-proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"cosmossdk.io/core/address"

	msgv1 "cosmossdk.io/api/cosmos/msg/v1"
)

// Context is a context for retrieving the list of signers from a
// message where signers are specified by the cosmos.msg.v1.signer protobuf
// option. It also contains the ProtoFileResolver and address.Codec's used
// for resolving message descriptors and converting addresses.
type Context struct {
	fileResolver          ProtoFileResolver
	typeResolver          protoregistry.MessageTypeResolver
	addressCodec          address.Codec
	validatorAddressCodec address.Codec
	getSignersFuncs       map[protoreflect.FullName]getSignersFunc
}

// Options are options for creating Context which will be used for signing operations.
type Options struct {
	// FileResolver is the protobuf file resolver to use for resolving message descriptors.
	// If it is nil, the global protobuf registry will be used.
	FileResolver ProtoFileResolver

	// TypeResolver is the protobuf type resolver to use for resolving message types.
	TypeResolver protoregistry.MessageTypeResolver

	// AddressCodec is the codec for converting addresses between strings and bytes.
	AddressCodec address.Codec

	// ValidatorAddressCodec is the codec for converting validator addresses between strings and bytes.
	ValidatorAddressCodec address.Codec
}

// ProtoFileResolver is a protodesc.Resolver that also allows iterating over all
// files descriptors. It is a subset of the methods supported by protoregistry.Files.
type ProtoFileResolver interface {
	protodesc.Resolver
	RangeFiles(func(protoreflect.FileDescriptor) bool)
}

// NewContext creates a new Context using the provided options.
func NewContext(options Options) (*Context, error) {
	protoFiles := options.FileResolver
	if protoFiles == nil {
		protoFiles = protoregistry.GlobalFiles
	}

	protoTypes := options.TypeResolver
	if protoTypes == nil {
		protoTypes = protoregistry.GlobalTypes
	}

	if options.AddressCodec == nil {
		return nil, errors.New("address codec is required")
	}

	if options.ValidatorAddressCodec == nil {
		return nil, errors.New("validator address codec is required")
	}

	c := &Context{
		fileResolver:          protoFiles,
		typeResolver:          protoTypes,
		addressCodec:          options.AddressCodec,
		validatorAddressCodec: options.ValidatorAddressCodec,
		getSignersFuncs:       map[protoreflect.FullName]getSignersFunc{},
	}

	return c, nil
}

type getSignersFunc func(proto.Message) ([][]byte, error)

func getSignersFieldNames(descriptor protoreflect.MessageDescriptor) ([]string, error) {
	signersFields := proto.GetExtension(descriptor.Options(), msgv1.E_Signer).([]string)
	if len(signersFields) == 0 {
		return nil, fmt.Errorf("no cosmos.msg.v1.signer option found for message %s", descriptor.FullName())
	}

	return signersFields, nil
}

// Validate performs a dry run of getting all msg's signers. This has 2 benefits:
// - it will error if any Msg has forgotten the "cosmos.msg.v1.signer"
// annotation
// - it will pre-populate the context's internal cache for getSignersFuncs
// so that calling it in antehandlers will be faster.
func (c *Context) Validate() error {
	var errs []error
	c.fileResolver.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		for i := 0; i < fd.Services().Len(); i++ {
			sd := fd.Services().Get(i)

			// Skip services that are not annotated with the "cosmos.msg.v1.service" option.
			if ext := proto.GetExtension(sd.Options(), msgv1.E_Service); ext == nil || !ext.(bool) {
				continue
			}

			for j := 0; j < sd.Methods().Len(); j++ {
				md := sd.Methods().Get(j).Input()
				_, err := c.getGetSignersFn(md)
				if err != nil && !errors.Is(err, NeedCustomSignersError) { // don't fail on custom signers
					errs = append(errs, err)
				}
			}
		}

		return true
	})

	return errors.Join(errs...)
}

func (c *Context) makeGetSignersFunc(descriptor protoreflect.MessageDescriptor) (getSignersFunc, error) {
	isCustom := proto.GetExtension(descriptor.Options(), msgv1.E_CustomSigner).(bool)
	if isCustom {
		return nil, fmt.Errorf("%w: %s", NeedCustomSignersError, descriptor.FullName())
	}

	signersFields, err := getSignersFieldNames(descriptor)
	if err != nil {
		return nil, err
	}

	fieldGetters := make([]func(proto.Message, [][]byte) ([][]byte, error), len(signersFields))
	for i, fieldName := range signersFields {
		field := descriptor.Fields().ByName(protoreflect.Name(fieldName))
		if field == nil {
			return nil, fmt.Errorf("field %s not found in message %s", fieldName, descriptor.FullName())
		}

		if field.IsMap() || field.HasOptionalKeyword() {
			return nil, fmt.Errorf("cosmos.msg.v1.signer field %s in message %s must not be a map or optional", fieldName, descriptor.FullName())
		}

		switch field.Kind() {
		case protoreflect.StringKind:
			addrCdc := c.getAddressCodec(field)
			if field.IsList() {
				fieldGetters[i] = func(msg proto.Message, arr [][]byte) ([][]byte, error) {
					signers := msg.ProtoReflect().Get(field).List()
					n := signers.Len()
					for i := 0; i < n; i++ {
						addrStr := signers.Get(i).String()
						addrBz, err := addrCdc.StringToBytes(addrStr)
						if err != nil {
							return nil, err
						}
						arr = append(arr, addrBz)
					}
					return arr, nil
				}
			} else {
				fieldGetters[i] = func(msg proto.Message, arr [][]byte) ([][]byte, error) {
					addrStr := msg.ProtoReflect().Get(field).String()
					addrBz, err := addrCdc.StringToBytes(addrStr)
					if err != nil {
						return nil, err
					}
					return append(arr, addrBz), nil
				}
			}
		case protoreflect.MessageKind:
			isList := field.IsList()
			nestedMessage := field.Message()
			nestedSignersFields, err := getSignersFieldNames(nestedMessage)
			if err != nil {
				return nil, err
			}

			if len(nestedSignersFields) != 1 {
				return nil, fmt.Errorf("nested cosmos.msg.v1.signer option in message %s must contain only one value", nestedMessage.FullName())
			}

			nestedFieldName := nestedSignersFields[0]
			nestedField := nestedMessage.Fields().ByName(protoreflect.Name(nestedFieldName))
			nestedIsList := nestedField.IsList()
			if nestedField == nil {
				return nil, fmt.Errorf("field %s not found in message %s", nestedFieldName, nestedMessage.FullName())
			}

			if nestedField.Kind() != protoreflect.StringKind || nestedField.IsMap() || nestedField.HasOptionalKeyword() {
				return nil, fmt.Errorf("nested signer field %s in message %s must be a simple string", nestedFieldName, nestedMessage.FullName())
			}

			addrCdc := c.getAddressCodec(nestedField)

			if isList {
				if nestedIsList {
					fieldGetters[i] = func(msg proto.Message, arr [][]byte) ([][]byte, error) {
						msgs := msg.ProtoReflect().Get(field).List()
						m := msgs.Len()
						for i := 0; i < m; i++ {
							signers := msgs.Get(i).Message().Get(nestedField).List()
							n := signers.Len()
							for j := 0; j < n; j++ {
								addrStr := signers.Get(j).String()
								addrBz, err := addrCdc.StringToBytes(addrStr)
								if err != nil {
									return nil, err
								}
								arr = append(arr, addrBz)
							}
						}
						return arr, nil
					}
				} else {
					fieldGetters[i] = func(msg proto.Message, arr [][]byte) ([][]byte, error) {
						msgs := msg.ProtoReflect().Get(field).List()
						m := msgs.Len()
						for i := 0; i < m; i++ {
							addrStr := msgs.Get(i).Message().Get(nestedField).String()
							addrBz, err := addrCdc.StringToBytes(addrStr)
							if err != nil {
								return nil, err
							}
							arr = append(arr, addrBz)
						}
						return arr, nil
					}
				}
			} else {
				if nestedIsList {
					fieldGetters[i] = func(msg proto.Message, arr [][]byte) ([][]byte, error) {
						nestedMsg := msg.ProtoReflect().Get(field).Message()
						signers := nestedMsg.Get(nestedField).List()
						n := signers.Len()
						for j := 0; j < n; j++ {
							addrStr := signers.Get(j).String()
							addrBz, err := addrCdc.StringToBytes(addrStr)
							if err != nil {
								return nil, err
							}
							arr = append(arr, addrBz)
						}
						return arr, nil
					}
				} else {
					fieldGetters[i] = func(msg proto.Message, arr [][]byte) ([][]byte, error) {
						addrStr := msg.ProtoReflect().Get(field).Message().Get(nestedField).String()
						addrBz, err := addrCdc.StringToBytes(addrStr)
						if err != nil {
							return nil, err
						}
						return append(arr, addrBz), nil
					}
				}
			}

		default:
			return nil, fmt.Errorf("unexpected field type %s for field %s in message %s", field.Kind(), fieldName, descriptor.FullName())
		}
	}

	return func(message proto.Message) ([][]byte, error) {
		var signers [][]byte
		for _, getter := range fieldGetters {
			signers, err = getter(message, signers)
			if err != nil {
				return nil, err
			}
		}
		return signers, nil
	}, nil
}

func (c *Context) getAddressCodec(field protoreflect.FieldDescriptor) address.Codec {
	scalarOpt := proto.GetExtension(field.Options(), cosmos_proto.E_Scalar)
	addrCdc := c.addressCodec
	if scalarOpt != nil {
		if scalarOpt.(string) == "cosmos.ValidatorAddressString" {
			addrCdc = c.validatorAddressCodec
		}
	}

	return addrCdc
}

func (c *Context) getGetSignersFn(messageDescriptor protoreflect.MessageDescriptor) (getSignersFunc, error) {
	f, ok := c.getSignersFuncs[messageDescriptor.FullName()]
	if !ok {
		var err error
		f, err = c.makeGetSignersFunc(messageDescriptor)
		if err != nil {
			return nil, err
		}
		c.getSignersFuncs[messageDescriptor.FullName()] = f
	}

	return f, nil
}

// GetSigners returns the signers for a given message.
func (c *Context) GetSigners(msg proto.Message) ([][]byte, error) {
	f, err := c.getGetSignersFn(msg.ProtoReflect().Descriptor())
	if err != nil {
		return nil, err
	}

	return f(msg)
}

// AddressCodec returns the address codec used by the context.
func (c *Context) AddressCodec() address.Codec {
	return c.addressCodec
}

// ValidatorAddressCodec returns the validator address codec used by the context.
func (c *Context) ValidatorAddressCodec() address.Codec {
	return c.validatorAddressCodec
}

// FileResolver returns the protobuf file resolver used by the context.
func (c *Context) FileResolver() ProtoFileResolver {
	return c.fileResolver
}

// TypeResolver returns the protobuf type resolver used by the context.
func (c *Context) TypeResolver() protoregistry.MessageTypeResolver {
	return c.typeResolver
}

// DefineCustomGetSigners defines a custom GetSigners function for a given
// message type. It is defined as a function rather than a method on Context
// because of how go generics work.
//
// NOTE: if a custom signers function is defined, the message type used to
// define this function MUST be the concrete type passed to GetSigners,
// otherwise a runtime type error will occur.
func DefineCustomGetSigners[T proto.Message](ctx *Context, getSigners func(T) ([][]byte, error)) {
	t := *new(T)
	ctx.getSignersFuncs[t.ProtoReflect().Descriptor().FullName()] = func(msg proto.Message) ([][]byte, error) {
		return getSigners(msg.(T))
	}
}

var NeedCustomSignersError = errors.New("need custom signers function")
