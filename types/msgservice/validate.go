package msgservice

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	msg "cosmossdk.io/api/cosmos/msg/v1"
)

// ValidateAnnotations validates that the proto annotations are correct.
// More specifically, it verifies:
// - all services named "Msg" have `(cosmos.msg.v1.service) = true`,
// - all sdk.Msgs have correct `(cosmos.msg.v1.signer) = "..."`.
//
// More validations can be added here in the future.
//
// If `protoFiles` is nil, then protoregistry.GlobalFile will be used.
func ValidateProtoAnnotations(protoFiles *protoregistry.Files) error {
	if protoFiles == nil {
		protoFiles = protoregistry.GlobalFiles
	}

	var (
		serviceErrs []error
		messageErrs []error
	)
	protoFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		for i := 0; i < fd.Services().Len(); i++ {
			sd := fd.Services().Get(i)
			if sd.Name() == "Msg" {
				// We use the heuristic that services name Msg are exactly the
				// ones that need the proto annotations check.
				err := validateMsgServiceAnnotations(protoFiles, sd)
				if err != nil {
					serviceErrs = append(serviceErrs, err)
				}

				for j := 0; j < sd.Methods().Len(); j++ {
					err := validateMsgAnnotations(protoFiles, sd.Methods().Get(j).Input())
					if err != nil {
						messageErrs = append(messageErrs, err)
					}
				}
			}
		}

		return true
	})

	return errors.Join(errors.Join(serviceErrs...), errors.Join(messageErrs...))
}

// validateMsgServiceAnnotations validates that the service has the
// `(cosmos.msg.v1.service) = true` proto annotation.
func validateMsgServiceAnnotations(protoFiles *protoregistry.Files, sd protoreflect.ServiceDescriptor) error {
	ext := proto.GetExtension(sd.Options(), msg.E_Service)
	isService, ok := ext.(bool)
	if !ok {
		return fmt.Errorf("expected bool, got %T", ext)
	}

	if !isService {
		return fmt.Errorf("service %s does not have cosmos.msg.v1.service proto annotation", sd.FullName())
	}

	return nil
}

// validateMsgAnnotations validates the sdk.Msg service has the correct
// `(cosmos.msg.v1.signer) = "..."` proto annotation.
func validateMsgAnnotations(protoFiles *protoregistry.Files, md protoreflect.MessageDescriptor) error {
	ext := proto.GetExtension(md.Options(), msg.E_Signer)
	signers, ok := ext.([]string)
	if !ok {
		return fmt.Errorf("expected []string, got %T", ext)
	}

	if len(signers) == 0 {
		return fmt.Errorf("sdk.Msg %s does not have cosmos.msg.v1.signer proto annotation", md.FullName())
	}

	for i, signer := range signers {
		if signer == "" {
			return fmt.Errorf("sdk.Msg %s signer at index %d is empty", md.FullName(), i)
		}

		// Make sure the signer annotation is a correct field of type string
		fd := md.Fields().ByName(protoreflect.Name(signer))
		if fd == nil {
			return fmt.Errorf("sdk.Msg %s has incorrect signer %s", md.FullName(), signer)
		}

		// The signer annotation should point to:
		// - either be a string field,
		// - or a message field who recursively has a "signer" string field.
		switch fd.Kind() {
		case protoreflect.StringKind:
			continue
		case protoreflect.MessageKind:
			err := validateMsgAnnotations(protoFiles, fd.Message())
			if err != nil {
				return err
			}

			continue
		default:
			return fmt.Errorf("sdk.Msg %s has signer %s of incorrect type; expected string or message, got %s", md.FullName(), signer, fd.Kind())
		}
	}

	return nil
}
