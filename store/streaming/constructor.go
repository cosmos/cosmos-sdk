package streaming

import (
	"fmt"
	"strings"
	"sync"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store/streaming/file"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/spf13/cast"
)

// ServiceConstructor is used to construct a streaming service
type ServiceConstructor func(opts serverTypes.AppOptions, keys []sdk.StoreKey, marshaller codec.BinaryCodec) (baseapp.StreamingService, error)

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
	if constructor, ok := ServiceConstructorLookupTable[ssType]; ok && constructor != nil {
		return constructor, nil
	}
	return nil, fmt.Errorf("streaming service constructor of type %s not found", ssType.String())
}

// FileStreamingConstructor is the streaming.ServiceConstructor function for creating a FileStreamingService
func FileStreamingConstructor(opts serverTypes.AppOptions, keys []sdk.StoreKey, marshaller codec.BinaryCodec) (baseapp.StreamingService, error) {
	filePrefix := cast.ToString(opts.Get("streamers.file.prefix"))
	fileDir := cast.ToString(opts.Get("streamers.file.writeDir"))
	if fileDir == "" {
		fileDir = file.DefaultWriteDir
	}
	return file.NewStreamingService(fileDir, filePrefix, keys, marshaller)
}

// LoadStreamingServices is a function for loading StreamingServices onto the BaseApp using the provided AppOptions, codec, and keys
// It returns the WaitGroup and quit channel used to synchronize with the streaming services and any error that occurs during the setup
func LoadStreamingServices(bApp *baseapp.BaseApp, appOpts serverTypes.AppOptions, appCodec codec.BinaryCodec, keys map[string]*sdk.KVStoreKey) (*sync.WaitGroup, chan struct{}, error) {
	// waitgroup and quit channel for optional shutdown coordination of the streaming service(s)
	wg := new(sync.WaitGroup)
	quitChan := make(chan struct{})
	// configure state listening capabilities using AppOptions
	streamers := cast.ToStringSlice(appOpts.Get("store.streamers"))
	for _, streamerName := range streamers {
		// get the store keys allowed to be exposed for this streaming service
		exposeKeyStrs := cast.ToStringSlice(appOpts.Get(fmt.Sprintf("streamers.%s.keys", streamerName)))
		exposeStoreKeys := make([]sdk.StoreKey, 0, len(exposeKeyStrs))
		for _, keyStr := range exposeKeyStrs {
			if storeKey, ok := keys[keyStr]; ok {
				exposeStoreKeys = append(exposeStoreKeys, storeKey)
			}
		}
		// get the constructor for this streamer name
		constructor, err := NewServiceConstructor(streamerName)
		if err != nil {
			// close the quitChan to shutdown any services we may have already spun up before hitting the error on this one
			close(quitChan)
			return nil, nil, err
		}
		// generate the streaming service using the constructor, appOptions, and the StoreKeys we want to expose
		streamingService, err := constructor(appOpts, exposeStoreKeys, appCodec)
		if err != nil {
			close(quitChan)
			return nil, nil, err
		}
		// register the streaming service with the BaseApp
		bApp.SetStreamingService(streamingService)
		// kick off the background streaming service loop
		streamingService.Stream(wg, quitChan)
	}
	return wg, quitChan, nil
}
