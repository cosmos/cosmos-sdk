package codec

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/dynamicpb"
)

var (
	ErrFileRegistered      = errors.New("file is already registered")
	ErrNoDependencyFetcher = errors.New("no dependency fetcher was set")
	ErrBuild               = errors.New("unable to build the file descriptor")
)

// Codec is a protobuf registry builder. It is capable of building
// entire protobuf registries starting from one file and resolving
// its imports dynamically. It's meant to be used in a contextualized
// way in order to avoid namespace issues caused by identical filenames
// package names and message names.
// It has support for *anypb.Any type resolving, as long as the types
// were correctly registered.
type Codec struct {
	files *protoregistry.Files
	types *protoregistry.Types

	jsonMarshaler   protojson.MarshalOptions
	jsonUnmarshaler protojson.UnmarshalOptions

	protoMarshaler   proto.MarshalOptions
	protoUnmarshaler proto.UnmarshalOptions

	dependencyFetcher ProtoImportsDownloader
}

// NewCodec builds a codec which resolves dependencies for unknown protobuf types
func NewCodec(f ProtoImportsDownloader) *Codec {
	filesReg := new(protoregistry.Files)
	typesReg := new(protoregistry.Types)

	typeResolver := newTypeResolver(typesReg)
	return &Codec{
		files: filesReg,
		types: typesReg,
		jsonMarshaler: protojson.MarshalOptions{
			Resolver: typeResolver,
		},
		jsonUnmarshaler: protojson.UnmarshalOptions{
			Resolver: typesReg,
		},
		protoMarshaler: proto.MarshalOptions{},
		protoUnmarshaler: proto.UnmarshalOptions{
			Resolver: typeResolver,
		},
		dependencyFetcher: f,
	}
}

func (c *Codec) KnownMessage(name string) bool {
	if _, err := c.types.FindMessageByName(protoreflect.FullName(name)); err != nil {
		return false
	}

	return true
}

func (c *Codec) MarshalJSON(o proto.Message) (b []byte, err error) {
	return c.jsonMarshaler.Marshal(o)
}

func (c *Codec) UnmarshalJSON(b []byte, o proto.Message) error {
	return c.jsonUnmarshaler.Unmarshal(b, o)
}

func (c *Codec) Marshal(o proto.Message) ([]byte, error) {
	return c.protoMarshaler.Marshal(o)
}

func (c *Codec) Unmarshal(b []byte, o proto.Message) error {
	return c.protoUnmarshaler.Unmarshal(b, o)
}

// FilesRegistry returns the codec proto file registry with read only access
func (c *Codec) FilesRegistry() ReadonlyProtoFileRegistry {
	return c.files
}

// TypesRegistry returns the codec proto type registry with read only access
func (c *Codec) TypesRegistry() ReadonlyTypeRegistry {
	return c.types
}

// RegisterRawFileDescriptor is going to parse the given descriptor and also attempt to resolve its import dependencies
func (c *Codec) RegisterRawFileDescriptor(ctx context.Context, rawDesc []byte) (fileDesc protoreflect.FileDescriptor, err error) {

	rawDesc, err = tryUnzip(rawDesc)
	if err != nil {
		return nil, err
	}
	// we build a temporary descriptor whose purpose is to check for dependencies
	// after the proto dependencies are resolved and registered we can build
	// and register the file descriptor
	tmpDesc, err := buildDescriptor(new(protoregistry.Files), new(protoregistry.Types), rawDesc)
	if err != nil {
		return nil, err
	}

	// get dependencies
	fileImports := tmpDesc.Imports()
	for i := 0; i < fileImports.Len(); i++ {
		// process missing imports
		imp := fileImports.Get(i)
		_, err := c.files.FindFileByPath(imp.Path())
		// if the file exist then skip the import
		if err == nil {
			continue
		}
		// if the error is not a not found one then fail
		if !errors.Is(err, protoregistry.NotFound) {
			return nil, fmt.Errorf("unrecognized error while processing imports: %s", err)
		}

		// check if we have set up a dependency fetcher
		if c.dependencyFetcher == nil {
			return nil, fmt.Errorf("file %s requires missing dependency %s: %w", tmpDesc.Path(), imp.Path(), ErrNoDependencyFetcher)
		}
		// get the missing import from the fetcher
		// TODO: files such as gogoproto and co are empty :\
		importDesc, err := c.dependencyFetcher.DownloadDescriptorByPath(ctx, imp.Path())
		if err != nil {
			return nil, fmt.Errorf("unable to fetch missing dependency for %s: %s: %w", tmpDesc.Path(), imp.Path(), err)
		}
		_, err = c.RegisterRawFileDescriptor(ctx, importDesc)
		if err != nil && !errors.Is(err, ErrFileRegistered) {
			return nil, fmt.Errorf("unable to parse missing dependency for %s: %s: %w", tmpDesc.Path(), imp.Path(), err)
		}
	}
	// after we have registered the dependencies
	// we rebuild the descriptor with the registry
	// that contains the resolved dependencies
	fileDesc, err = buildDescriptor(c.files, c.types, rawDesc)
	if err != nil {
		return fileDesc, err
	}
	return fileDesc, err
}

// tryUnzip is going to attempt to unzip the provided descriptor
func tryUnzip(rawDesc []byte) ([]byte, error) {
	buf := bytes.NewBuffer(rawDesc)
	r, err := gzip.NewReader(buf)
	if err != nil {
		return rawDesc, nil
	}

	unzipped, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return unzipped, nil
}

// buildDescriptor builds a proto file descriptor given the decoded bytes
func buildDescriptor(fileRegistry *protoregistry.Files, typesRegistry *protoregistry.Types, rawDesc []byte) (fileDesc protoreflect.FileDescriptor, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %#v", ErrBuild, r)
		}
	}()

	tmpBuilder := (&protoimpl.DescBuilder{
		GoPackagePath: "",
		RawDescriptor: rawDesc,
		TypeResolver:  new(protoregistry.Types),
		FileRegistry:  new(protoregistry.Files),
	}).Build()

	filePath := tmpBuilder.File.Path()

	// check if it exists
	existingFd, err := fileRegistry.FindFileByPath(filePath)
	if err != nil && errors.Is(err, protoregistry.NotFound) {
		// does not exist, build file desc
		fd := (&protoimpl.DescBuilder{
			RawDescriptor: rawDesc,
			TypeResolver:  typesRegistry,
			FileRegistry:  fileRegistry,
		}).Build().File
		// add fd types to registry
		err = regTypes(typesRegistry, fd)
		if err != nil {
			return nil, err
		}
		return fd, err
	}
	// check if err
	if err != nil {
		return nil, err
	}
	// already registered
	return existingFd, fmt.Errorf("%s: %w", existingFd.Path(), ErrFileRegistered)
}

func regTypes(reg *protoregistry.Types, fd protoreflect.FileDescriptor) error {
	msgs := fd.Messages()
	for i := 0; i < msgs.Len(); i++ {
		md := msgs.Get(i)
		typ := dynamicpb.NewMessageType(md)
		err := reg.RegisterMessage(typ)
		if err != nil {
			return err
		}
	}

	enums := fd.Enums()
	for i := 0; i < enums.Len(); i++ {
		ed := enums.Get(i)
		typ := dynamicpb.NewEnumType(ed)
		err := reg.RegisterEnum(typ)
		if err != nil {
			return err
		}
	}

	extensions := fd.Extensions()
	for i := 0; i < extensions.Len(); i++ {
		xd := extensions.Get(i)
		typ := dynamicpb.NewExtensionType(xd)
		err := reg.RegisterExtension(typ)
		if err != nil {
			return err
		}
	}

	return nil
}

type typeResolver struct {
	reg *protoregistry.Types
}

func (t typeResolver) FindExtensionByName(field protoreflect.FullName) (protoreflect.ExtensionType, error) {
	return t.reg.FindExtensionByName(field)
}

func (t typeResolver) FindExtensionByNumber(message protoreflect.FullName, field protoreflect.FieldNumber) (protoreflect.ExtensionType, error) {
	return t.reg.FindExtensionByNumber(message, field)
}

func (t typeResolver) FindMessageByName(message protoreflect.FullName) (protoreflect.MessageType, error) {
	return t.reg.FindMessageByName(message)
}

func (t typeResolver) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	u := url
	if strings.HasPrefix(u, "/") {
		u = u[1:]
	}
	return t.reg.FindMessageByURL(u)
}

func newTypeResolver(reg *protoregistry.Types) *typeResolver {
	return &typeResolver{reg: reg}
}
