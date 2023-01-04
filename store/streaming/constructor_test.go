package streaming_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store/streaming"
	"github.com/cosmos/cosmos-sdk/store/streaming/file"
	"github.com/cosmos/cosmos-sdk/store/types"
)

type fakeOptions struct{}

func (f *fakeOptions) Get(key string) interface{} {
	if key == "streamers.file.write_dir" {
		return "data/file_streamer"
	}
	return nil
}

var (
	mockOptions    = new(fakeOptions)
	mockKeys       = []types.StoreKey{types.NewKVStoreKey("mockKey1"), types.NewKVStoreKey("mockKey2")}
	testMarshaller = types.NewTestCodec()
)

func TestStreamingServiceConstructor(t *testing.T) {
	_, err := streaming.NewServiceConstructor("unexpectedName")
	require.NotNil(t, err)

	constructor, err := streaming.NewServiceConstructor("file")
	require.Nil(t, err)
	var expectedType streaming.ServiceConstructor
	require.IsType(t, expectedType, constructor)

	serv, err := constructor(mockOptions, mockKeys, testMarshaller, log.NewNopLogger())
	require.Nil(t, err)
	require.IsType(t, &file.StreamingService{}, serv)
	listeners := serv.Listeners()
	for _, key := range mockKeys {
		_, ok := listeners[key]
		require.True(t, ok)
	}
}

func TestLoadStreamingServices(t *testing.T) {
	encCdc := types.NewTestCodec()
	keys := types.NewKVStoreKeys("mockKey1", "mockKey2")

	testCases := map[string]struct {
		appOpts            serverTypes.AppOptions
		activeStreamersLen int
	}{
		"empty app options": {
			appOpts: emptyAppOptions{},
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
			activeStreamers, _, err := streaming.LoadStreamingServices(tc.appOpts, encCdc, log.NewNopLogger(), keys)
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
	case "streamers.file.write_dir":
		return "data/file_streamer"
	default:
		return nil
	}
}

type emptyAppOptions struct{}

func (ao emptyAppOptions) Get(o string) interface{} {
	return nil
}
