package streaming_test

import (
	"testing"

	"github.com/pointnetwork/cosmos-point-sdk/baseapp"
	"github.com/pointnetwork/cosmos-point-sdk/codec"
	codecTypes "github.com/pointnetwork/cosmos-point-sdk/codec/types"
	serverTypes "github.com/pointnetwork/cosmos-point-sdk/server/types"
	"github.com/pointnetwork/cosmos-point-sdk/simapp"
	"github.com/pointnetwork/cosmos-point-sdk/store/streaming"
	"github.com/pointnetwork/cosmos-point-sdk/store/streaming/file"
	"github.com/pointnetwork/cosmos-point-sdk/store/types"
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/stretchr/testify/require"
)

type fakeOptions struct{}

func (f *fakeOptions) Get(string) interface{} { return nil }

var (
	mockOptions       = new(fakeOptions)
	mockKeys          = []types.StoreKey{sdk.NewKVStoreKey("mockKey1"), sdk.NewKVStoreKey("mockKey2")}
	interfaceRegistry = codecTypes.NewInterfaceRegistry()
	testMarshaller    = codec.NewProtoCodec(interfaceRegistry)
)

func TestStreamingServiceConstructor(t *testing.T) {
	_, err := streaming.NewServiceConstructor("unexpectedName")
	require.NotNil(t, err)

	constructor, err := streaming.NewServiceConstructor("file")
	require.Nil(t, err)
	var expectedType streaming.ServiceConstructor
	require.IsType(t, expectedType, constructor)

	serv, err := constructor(mockOptions, mockKeys, testMarshaller)
	require.Nil(t, err)
	require.IsType(t, &file.StreamingService{}, serv)
	listeners := serv.Listeners()
	for _, key := range mockKeys {
		_, ok := listeners[key]
		require.True(t, ok)
	}
}

func TestLoadStreamingServices(t *testing.T) {
	db := dbm.NewMemDB()
	encCdc := simapp.MakeTestEncodingConfig()
	keys := sdk.NewKVStoreKeys("mockKey1", "mockKey2")
	bApp := baseapp.NewBaseApp("appName", log.NewNopLogger(), db, nil)

	testCases := map[string]struct {
		appOpts            serverTypes.AppOptions
		activeStreamersLen int
	}{
		"empty app options": {
			appOpts: simapp.EmptyAppOptions{},
		},
		"all StoreKeys exposed": {
			appOpts:            streamingAppOptions{keys: []string{"*"}},
			activeStreamersLen: 1,
		},
		"some StoreKey exposed": {
			appOpts:            streamingAppOptions{keys: []string{"mockKey1"}},
			activeStreamersLen: 1,
		},
		"not exposing anything": {
			appOpts: streamingAppOptions{keys: []string{"mockKey3"}},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			activeStreamers, _, err := streaming.LoadStreamingServices(bApp, tc.appOpts, encCdc.Codec, keys)
			require.NoError(t, err)
			require.Equal(t, tc.activeStreamersLen, len(activeStreamers))
		})
	}
}

type streamingAppOptions struct {
	keys []string
}

func (ao streamingAppOptions) Get(o string) interface{} {
	switch o {
	case "store.streamers":
		return []string{"file"}
	case "streamers.file.keys":
		return ao.keys
	default:
		return nil
	}
}
