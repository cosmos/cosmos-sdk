package codec

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoimpl"
)

var (
	ErrFileRegistered      = errors.New("file is already registered")
	ErrNoDependencyFetcher = errors.New("no dependency fetcher was set")
	ErrBuild               = errors.New("unable to build the file descriptor")
)

// Registry is a recursive proto dependencies resolver
// if instantiated with NewFetcherRegistry it will
// resolve proto imports
type Registry struct {
	files *protoregistry.Files
	types *protoregistry.Types

	dependencyFetcher DependencyFetcher
}

func NewRegistry() *Registry {
	return &Registry{
		files: &protoregistry.Files{},
		types: &protoregistry.Types{},
	}
}

// NewFetcherRegistry builds a registry which also
// resolves protobuf dependencies
func NewFetcherRegistry(f DependencyFetcher) *Registry {
	return &Registry{
		files:             new(protoregistry.Files),
		types:             new(protoregistry.Types),
		dependencyFetcher: f,
	}
}

func (b *Registry) ImportedFiles() []string {
	var resp []string
	b.files.RangeFiles(func(descriptor protoreflect.FileDescriptor) bool {
		resp = append(resp, descriptor.Path())
		return true
	})

	return resp
}

// Parse is going to parse the given descriptor
// and also attempt to resolve its import dependencies
func (b *Registry) Parse(rawDesc []byte) (fileDesc protoreflect.FileDescriptor, err error) {

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
		_, err := b.files.FindFileByPath(imp.Path())
		// if the file exist then skip the import
		if err == nil {
			continue
		}
		// if the error is not a not found one then fail
		if !errors.Is(err, protoregistry.NotFound) {
			return nil, fmt.Errorf("unrecognized error while processing imports: %s", err)
		}

		// check if we have set up a dependency fetcher
		if b.dependencyFetcher == nil {
			return nil, fmt.Errorf("file %s requires missing dependency %s: %w", tmpDesc.Path(), imp.Path(), ErrNoDependencyFetcher)
		}
		// get the missing import from the fetcher
		// TODO: files such as gogoproto and co are empty :\
		importDesc, err := b.dependencyFetcher.Fetch(imp.Path())
		if err != nil {
			return nil, fmt.Errorf("unable to fetch missing dependency for %s: %s: %w", tmpDesc.Path(), imp.Path(), err)
		}
		_, err = b.Parse(importDesc)
		if err != nil && !errors.Is(err, ErrFileRegistered) {
			return nil, fmt.Errorf("unable to parse missing dependency for %s: %s: %w", tmpDesc.Path(), imp.Path(), err)
		}
	}
	// after we have registered the dependencies
	// we rebuild the descriptor with the registry
	// that contains the resolved dependencies
	fileDesc, err = buildDescriptor(b.files, b.types, rawDesc)
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
		return (&protoimpl.DescBuilder{
			RawDescriptor: rawDesc,
			TypeResolver:  typesRegistry,
			FileRegistry:  fileRegistry,
		}).Build().File, nil
	}
	// check if err
	if err != nil {
		return nil, err
	}
	// already registered
	return existingFd, fmt.Errorf("%s: %w", existingFd.Path(), ErrFileRegistered)
}
