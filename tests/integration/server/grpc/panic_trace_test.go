package grpc_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"cosmossdk.io/depinject"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/v2/pruning/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

type panickingQueryServer struct {
	testdata.QueryImpl
}

func (panickingQueryServer) Echo(context.Context, *testdata.EchoRequest) (*testdata.EchoResponse, error) {
	panic("query handler panicked")
}

func TestGRPCQueryPanicRespectsTrace(t *testing.T) {
	for _, trace := range []bool{false, true} {
		if !t.Run(fmt.Sprintf("trace=%t", trace), func(t *testing.T) {
			appConfig := network.MinimumAppConfig()
			cfg, err := network.DefaultConfigWithAppConfig(appConfig)
			require.NoError(t, err)
			cfg.NumValidators = 1
			cfg.AppConstructor = func(val network.ValidatorI) servertypes.Application {
				var appBuilder *runtime.AppBuilder
				err := depinject.Inject(
					depinject.Configs(appConfig, depinject.Supply(val.GetCtx().Logger)),
					&appBuilder,
				)
				if err != nil {
					panic(err)
				}
				app := appBuilder.Build(
					dbm.NewMemDB(),
					baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
					baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
					baseapp.SetChainID(cfg.ChainID),
					baseapp.SetQueryGasLimit(0),
					baseapp.SetTrace(trace),
				)
				testdata.RegisterQueryServer(app.GRPCQueryRouter(), panickingQueryServer{})
				if err := app.Load(true); err != nil {
					panic(err)
				}
				return app
			}

			netw, err := network.New(t, t.TempDir(), cfg)
			require.NoError(t, err)
			t.Cleanup(netw.Cleanup)
			_, err = netw.WaitForHeightWithTimeout(2, 30*time.Second)
			require.NoError(t, err)

			val := netw.Validators[0]
			_, port, err := net.SplitHostPort(val.GetAppConfig().GRPC.Address)
			require.NoError(t, err)
			conn, err := grpc.NewClient(
				net.JoinHostPort("127.0.0.1", port),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(cfg.InterfaceRegistry).GRPCCodec())),
			)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, conn.Close()) })
			queryClient := testdata.NewQueryClient(val.ClientCtx.WithGRPCClient(conn))

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, "1")
			_, err = queryClient.Echo(ctx, &testdata.EchoRequest{}, grpc.WaitForReady(true))
			require.Equal(t, codes.Internal, status.Code(err))
			message := status.Convert(err).Message()
			if trace {
				require.Contains(t, message, "query handler panicked")
				require.Contains(t, message, "panickingQueryServer")
				require.Contains(t, message, "panic_trace_test.go:")
			} else {
				require.Equal(t, "query handler panicked", message)
				require.EqualError(t, err, "rpc error: code = Internal desc = query handler panicked")
			}

			ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, "1")
			response, err := queryClient.SayHello(ctx, &testdata.SayHelloRequest{Name: "after panic"}, grpc.WaitForReady(true))
			require.NoError(t, err)
			require.Equal(t, "Hello after panic!", response.Greeting)
		}) {
			// Startup failures can leave the network lock held, so do not start another network.
			return
		}
	}
}
