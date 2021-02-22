package streaming

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/streaming/file"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/spf13/cast"
)

// ServiceConstructor is used to construct a streaming service
type ServiceConstructor func(opts serverTypes.AppOptions, keys []sdk.StoreKey, marshaller codec.BinaryMarshaler) (sdk.StreamingService, error)

// ServiceType enum for specifying the type of StreamingService
type ServiceType int

const (
	Unknown ServiceType = iota
	File
	// add more in the future
)

// NewStreamingServiceType returns the streaming.ServiceType corresponding to the provided name
func NewStreamingServiceType(name string) ServiceType {
	switch strings.ToLower(name) {
	case "file", "f":
		return File
	default:
		return Unknown
	}
}

// String returns the string name of a streaming.ServiceType
func (sst ServiceType) String() string {
	switch sst {
	case File:
		return "file"
	default:
		return ""
	}
}

// ServiceConstructorLookupTable is a mapping of streaming.ServiceTypes to streaming.ServiceConstructors
var ServiceConstructorLookupTable = map[ServiceType]ServiceConstructor{
	File: FileStreamingConstructor,
}

// NewServiceConstructor returns the streaming.ServiceConstructor corresponding to the provided name
func NewServiceConstructor(name string) (ServiceConstructor, error) {
	ssType := NewStreamingServiceType(name)
	if ssType == Unknown {
		return nil, fmt.Errorf("unrecognized streaming service name %s", name)
	}
	if constructor, ok := ServiceConstructorLookupTable[ssType]; ok {
		return constructor, nil
	}
	return nil, fmt.Errorf("streaming service constructor of type %s not found", ssType.String())
}

// FileStreamingConstructor is the streaming.ServiceConstructor function for creating a FileStreamingService
func FileStreamingConstructor(opts serverTypes.AppOptions, keys []sdk.StoreKey, marshaller codec.BinaryMarshaler) (sdk.StreamingService, error) {
	filePrefix := cast.ToString(opts.Get("streamers.file.prefix"))
	fileDir := cast.ToString(opts.Get("streamers.file.writeDir"))
	return file.NewStreamingService(fileDir, filePrefix, keys, marshaller)
}
