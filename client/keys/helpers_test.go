package keys

import (
	"testing"

	"github.com/stretchr/testify/require"

	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/codec"
	_ "github.com/cosmos/cosmos-sdk/runtime"
)

func makeTestCodec(t *testing.T) codec.Codec {
	config := appconfig.Compose(&appv1alpha1.Config{
		Modules: []*appv1alpha1.ModuleConfig{
			{
				Name: "runtime",
				Config: appconfig.WrapAny(&runtimev1alpha1.Module{
					AppName: "clientTest",
				}),
			},
		},
	})

	var cdc codec.Codec
	err := depinject.Inject(config, &cdc)
	require.NoError(t, err)
	return cdc
}
