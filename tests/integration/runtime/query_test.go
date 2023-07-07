package runtime

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"gotest.tools/v3/assert"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
)

type fixture struct {
	ctx               sdk.Context
	appQueryClient    appv1alpha1.QueryClient
	autocliInfoClient autocliv1.QueryClient
	reflectionClient  reflectionv1.ReflectionServiceClient
}

func initFixture(t assert.TestingT) *fixture {
	f := &fixture{}

	var interfaceRegistry codectypes.InterfaceRegistry

	app, err := simtestutil.Setup(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AuthModule(),
				configurator.TxModule(),
				configurator.ParamsModule(),
				configurator.ConsensusModule(),
				configurator.BankModule(),
				configurator.StakingModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		&interfaceRegistry,
	)
	assert.NilError(t, err)

	f.ctx = app.BaseApp.NewContext(false)
	queryHelper := &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: app.BaseApp.GRPCQueryRouter(),
		Ctx:             f.ctx,
	}
	f.appQueryClient = appv1alpha1.NewQueryClient(queryHelper)
	f.autocliInfoClient = autocliv1.NewQueryClient(queryHelper)
	f.reflectionClient = reflectionv1.NewReflectionServiceClient(queryHelper)

	return f
}

func TestQueryAppConfig(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	res, err := f.appQueryClient.Config(f.ctx, &appv1alpha1.QueryConfigRequest{})
	assert.NilError(t, err)
	// app config is not nil
	assert.Assert(t, res != nil && res.Config != nil)

	moduleConfigs := map[string]*appv1alpha1.ModuleConfig{}
	for _, module := range res.Config.Modules {
		moduleConfigs[module.Name] = module
	}

	// has all expected modules
	for _, modName := range []string{"auth", "bank", "tx", "consensus", "runtime", "params", "staking"} {
		modConfig := moduleConfigs[modName]
		if modConfig == nil {
			t.Fatalf("missing %s", modName)
		}
		assert.Assert(t, modConfig.Config != nil)
	}
}

func TestReflectionService(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	res, err := f.reflectionClient.FileDescriptors(f.ctx, &reflectionv1.FileDescriptorsRequest{})
	assert.NilError(t, err)
	assert.Assert(t, res != nil && res.Files != nil)

	fdMap := map[string]*descriptorpb.FileDescriptorProto{}
	for _, descriptorProto := range res.Files {
		fdMap[*descriptorProto.Name] = descriptorProto
	}

	// check all file descriptors from gogo are present
	for path := range proto.AllFileDescriptors() {
		if fdMap[path] == nil {
			t.Fatalf("missing %s", path)
		}
	}

	// check all file descriptors from protoregistry are present
	protoregistry.GlobalFiles.RangeFiles(func(fileDescriptor protoreflect.FileDescriptor) bool {
		path := fileDescriptor.Path()
		if fdMap[path] == nil {
			t.Fatalf("missing %s", path)
		}
		return true
	})
}

func TestQueryAutoCLIAppOptions(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	res, err := f.autocliInfoClient.AppOptions(f.ctx, &autocliv1.AppOptionsRequest{})
	assert.NilError(t, err)
	assert.Assert(t, res != nil && res.ModuleOptions != nil)

	// make sure we have x/auth autocli options which were configured manually
	authOpts := res.ModuleOptions["auth"]
	assert.Assert(t, authOpts != nil)
	assert.Assert(t, authOpts.Query != nil)
	assert.Equal(t, "cosmos.auth.v1beta1.Query", authOpts.Query.Service)
	// make sure we have some custom options
	assert.Assert(t, len(authOpts.Query.RpcCommandOptions) != 0)

	// make sure we have x/staking autocli options which should have been auto-discovered
	stakingOpts := res.ModuleOptions["staking"]
	assert.Assert(t, stakingOpts != nil)
	assert.Assert(t, stakingOpts.Query != nil && stakingOpts.Tx != nil)
	assert.Equal(t, "cosmos.staking.v1beta1.Query", stakingOpts.Query.Service)
	assert.Equal(t, "cosmos.staking.v1beta1.Msg", stakingOpts.Tx.Service)

	// make sure tx module has no autocli options because it has no services
	assert.Assert(t, res.ModuleOptions["tx"] == nil)
}
