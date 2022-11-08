package msgservice

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	msg "cosmossdk.io/api/cosmos/msg/v1"

	// Import the descriptors
	_ "cosmossdk.io/api/cosmos/auth/v1beta1"
	_ "cosmossdk.io/api/cosmos/authz/v1beta1"
	_ "cosmossdk.io/api/cosmos/bank/v1beta1"
	_ "cosmossdk.io/api/cosmos/consensus/v1"
	_ "cosmossdk.io/api/cosmos/crisis/v1beta1"
	_ "cosmossdk.io/api/cosmos/distribution/v1beta1"
	_ "cosmossdk.io/api/cosmos/evidence/v1beta1"
	_ "cosmossdk.io/api/cosmos/feegrant/v1beta1"
	_ "cosmossdk.io/api/cosmos/gov/v1"
	_ "cosmossdk.io/api/cosmos/gov/v1beta1"
	_ "cosmossdk.io/api/cosmos/group/v1"
	_ "cosmossdk.io/api/cosmos/mint/v1beta1"
	_ "cosmossdk.io/api/cosmos/nft/v1beta1"
	_ "cosmossdk.io/api/cosmos/slashing/v1beta1"
	_ "cosmossdk.io/api/cosmos/staking/v1beta1"
	_ "cosmossdk.io/api/cosmos/upgrade/v1beta1"
	_ "cosmossdk.io/api/cosmos/vesting/v1beta1"
)

// ValidateServiceAnnotations validates that all Msg services have the
// `(cosmos.msg.v1.service) = true` proto annotation.
func ValidateServiceAnnotations(serviceName string) error {
	sd, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(serviceName))
	if err != nil {
		// If we don't find the pulsar-generated descriptors, we just skip
		// validation. This allows chain developers to migrate to pulsar at
		// their own pace.
		return nil
	}

	ext := proto.GetExtension(sd.Options(), msg.E_Service)
	hasOption, ok := ext.(bool)
	if !ok {
		return fmt.Errorf("expected bool, got %T", ext)
	}

	if !hasOption {
		return fmt.Errorf("service %s does not have cosmos.msg.v1.service proto annotation", serviceName)
	}

	return nil
}

// ValidateMsgAnnotations validates that all sdk.Msg services have the correct
// `(cosmos.msg.v1.signer) = "..."` proto annotation.
func ValidateMsgAnnotations(fqName string) error {
	d, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(fqName))
	if err != nil {
		// If we don't find the pulsar-generated descriptors, we just skip
		// validation. This allows chain developers to migrate to pulsar at
		// their own pace.
		return nil
	}
	md := d.(protoreflect.MessageDescriptor)

	ext := proto.GetExtension(md.Options(), msg.E_Signer)
	signers, ok := ext.([]string)
	if !ok {
		return fmt.Errorf("expected bool, got %T", ext)
	}

	if len(signers) == 0 {
		return fmt.Errorf("sdk.Msg %s does not have cosmos.msg.v1.signer proto annotation", fqName)
	}

	for i, signer := range signers {
		if signer == "" {
			return fmt.Errorf("sdk.Msg %s signer at index %d is empty", fqName, i)
		}

		// Make sure the signer annotation is a correct field of type string
		fd := md.Fields().ByName(protoreflect.Name(signer))
		if fd == nil {
			return fmt.Errorf("sdk.Msg %s has incorrect signer %s", fqName, signer)
		}

		if fd.Kind() != protoreflect.StringKind {
			return fmt.Errorf("sdk.Msg %s has signer %s of incorrect type; expected string, got %s", fqName, signer, fd.Kind())
		}
	}

	return nil
}
