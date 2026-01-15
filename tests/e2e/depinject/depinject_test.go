package depinject

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/address"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

var _ address.Codec = (*customAddressCodec)(nil)

type customAddressCodec struct{}

func (c customAddressCodec) StringToBytes(text string) ([]byte, error) {
	return []byte(text), nil
}

func (c customAddressCodec) BytesToString(bz []byte) (string, error) {
	return string(bz), nil
}

func TestAddressCodecFactory(t *testing.T) {
	var addrCodec address.Codec
	var valAddressCodec runtime.ValidatorAddressCodec
	var consAddressCodec runtime.ConsensusAddressCodec

	err := depinject.Inject(
		depinject.Configs(
			network.MinimumAppConfig(),
			depinject.Supply(log.NewNopLogger()),
		),
		&addrCodec, &valAddressCodec, &consAddressCodec)
	require.NoError(t, err)
	require.NotNil(t, addrCodec)
	_, ok := addrCodec.(customAddressCodec)
	require.False(t, ok)
	require.NotNil(t, valAddressCodec)
	_, ok = valAddressCodec.(customAddressCodec)
	require.False(t, ok)
	require.NotNil(t, consAddressCodec)
	_, ok = consAddressCodec.(customAddressCodec)
	require.False(t, ok)

	// Set the address codec to the custom one
	err = depinject.Inject(
		depinject.Configs(
			network.MinimumAppConfig(),
			depinject.Supply(
				log.NewNopLogger(),
				func() address.Codec { return customAddressCodec{} },
				func() runtime.ValidatorAddressCodec { return customAddressCodec{} },
				func() runtime.ConsensusAddressCodec { return customAddressCodec{} },
			),
		),
		&addrCodec, &valAddressCodec, &consAddressCodec)
	require.NoError(t, err)
	require.NotNil(t, addrCodec)
	_, ok = addrCodec.(customAddressCodec)
	require.True(t, ok)
	require.NotNil(t, valAddressCodec)
	_, ok = valAddressCodec.(customAddressCodec)
	require.True(t, ok)
	require.NotNil(t, consAddressCodec)
	_, ok = consAddressCodec.(customAddressCodec)
	require.True(t, ok)
}
