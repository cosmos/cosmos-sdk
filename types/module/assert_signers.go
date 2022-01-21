package module

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/api/cosmos/msg/v1"
	gogogrpc "github.com/gogo/protobuf/grpc"
	gogoproto "github.com/gogo/protobuf/proto"
	legacyproto "github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"io/ioutil"
)

var errBadSigners = errors.New("bad signers descriptor") // sentinel error for the panic message helpers.

// msgServerAssertSigners wraps a grpc.Server to check
// if registered msg server services inputs
// correctly signal message signers.
type msgServerAssertSigners struct {
	// importRemap exists because devs make the common mistake with proto imports
	// in which they don't import them as the codegen registers them.
	// See: https://github.com/cosmos/cosmos-sdk/issues/10978#issuecomment-1016644826
	importRemap map[string]string
	s           gogogrpc.Server
	files       *protoregistry.Files
}

// RegisterService registers a grpc.ServiceDesc to its real implementation
// after asserting that the inputs have the message signer extension.
func (s *msgServerAssertSigners) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	err := s.checkInputs(sd)
	if err != nil {
		s.onError(err)
	}

	s.s.RegisterService(sd, ss)
}

func (s *msgServerAssertSigners) checkInputs(sd *grpc.ServiceDesc) error {
	fd, err := s.FindFileByPath(sd.Metadata.(string))
	if err != nil {
		return err
	}

	prefSd := fd.Services().ByName(protoreflect.FullName(sd.ServiceName).Name())
	for i := 0; i < prefSd.Methods().Len(); i++ {
		md := prefSd.Methods().Get(i).Input()
		err := assertSigners(md, map[protoreflect.FullName]struct{}{})
		if err != nil {
			return fmt.Errorf("%w: %s", errBadSigners, err)
		}
	}

	return nil
}

func assertSigners(md protoreflect.MessageDescriptor, visited map[protoreflect.FullName]struct{}) error {
	if _, exists := visited[md.FullName()]; exists {
		return fmt.Errorf("recursive message with no signer: %#v", visited)
	}

	visited[md.FullName()] = struct{}{}

	signers := proto.GetExtension(md.Options(), msgv1.E_Signer).([]string)
	if len(signers) == 0 {
		return fmt.Errorf("no signers specified for cosmos message %s", md.FullName())
	}

	for _, signer := range signers {
		fd := md.Fields().ByName(protoreflect.Name(signer))
		if fd == nil {
			return fmt.Errorf("cosmos message %s signals %s as a signer field, but the field is not present", md.FullName(), signer)
		}

		switch fd.Kind() {
		case protoreflect.StringKind:
			return nil
		case protoreflect.MessageKind:
			err := assertSigners(fd.Message(), visited)
			if err != nil {
				return fmt.Errorf("cosmos message %s signals %s as signer field of message kind but assertion failed: %w", md.FullName(), fd.Name(), err)
			}
			return nil
		default:
			return fmt.Errorf("cosmos message %s signals %s as signer field but the field is not of kind message or string: %s", md.FullName(), fd.Name(), fd.Kind())
		}
	}

	return nil
}

func (s *msgServerAssertSigners) onError(err error) {
	const notFoundHelper = `The protobuf file registry did not find the given file, the reason for which this might be happening
is because you might have protobuf generated file which registers itself in the registry with a path but was then imported in other protobuf files
with a different path. Please use the github.com/cosmos/cosmos-sdk/types/module.WithProtoImportsRemaps function to rewrite the import paths.
Ref: https://github.com/cosmos/cosmos-sdk/issues/10978#issuecomment-1016644826
`
	const badSignersHelper = `Signers of the message were signaled in an incorrect way.
A signer field can be either of type string. Or it can be of type message. In case it is of type message that message must be extended with the cosmos.msg.v1.signer option.
For a string field example definition refer to: https://github.com/cosmos/cosmos-sdk/blob/fdymylja/agnostic-signers/proto/cosmos/bank/v1beta1/tx.proto#L23
For a message field example definition refer to: https://github.com/cosmos/cosmos-sdk/blob/fdymylja/agnostic-signers/proto/cosmos/bank/v1beta1/tx.proto#L39
`
	switch {
	case errors.Is(err, errBadSigners):
		err = fmt.Errorf("%w\n%s", err, badSignersHelper)
	case errors.Is(err, protoregistry.NotFound):
		err = fmt.Errorf("%w\n%s", err, notFoundHelper)
	}

	panic(err)
}

func newSignerChecker(server gogogrpc.Server, importRemap map[string]string) *msgServerAssertSigners {
	return &msgServerAssertSigners{
		importRemap: importRemap,
		s:           server,
		files:       new(protoregistry.Files),
	}
}

func (s *msgServerAssertSigners) FindFileByPath(path string) (protoreflect.FileDescriptor, error) {
	// we check if we need to do an import remapping
	originalPath := path
	if remap, exists := s.importRemap[path]; exists {
		path = remap
	}
	// first we check the populated registry
	fd, err := s.files.FindFileByPath(path)
	if err == nil {
		return fd, err
	}
	if !errors.Is(err, protoregistry.NotFound) {
		return nil, err
	}

	// we check the gogoproto registry, as we expect it to contain
	// most files
	fdZippedBytes := gogoproto.FileDescriptor(path)
	// if we don't find anything it means we might be trying to access
	// some wellknown protofile (ex: descriptor.proto) which is in
	// the legacy proto registry.
	if fdZippedBytes == nil {
		// nolint: staticcheck
		fdZippedBytes = legacyproto.FileDescriptor(path)
	}
	if fdZippedBytes == nil {
		return nil, fmt.Errorf("%s: %w", originalPath, protoregistry.NotFound)
	}

	fdBytes, err := unzip(fdZippedBytes)
	if err != nil {
		return nil, err
	}

	desc := &descriptorpb.FileDescriptorProto{}
	err = proto.Unmarshal(fdBytes, desc)
	if err != nil {
		return nil, err
	}

	fd, err = protodesc.NewFile(desc, s)
	if err != nil {
		return nil, err
	}

	err = s.files.RegisterFile(fd)
	if err != nil {
		return nil, err
	}

	return fd, nil
}

func (s *msgServerAssertSigners) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {
	// this should never be called.
	return s.files.FindDescriptorByName(name)
}

func unzip(b []byte) ([]byte, error) {
	if b == nil {
		return nil, nil
	}
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	unzipped, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return unzipped, nil
}
