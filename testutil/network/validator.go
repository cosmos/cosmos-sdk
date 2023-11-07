package network

import (
	"context"
	"net/http"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/cometbft/cometbft/node"
	cmtclient "github.com/cometbft/cometbft/rpc/client"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validator defines an in-process CometBFT validator node. Through this object,
// a client can make RPC and API calls and interact with any client command
// or handler.
type Validator struct {
	AppConfig  *srvconfig.Config
	ClientCtx  client.Context
	Ctx        *server.Context
	Dir        string
	NodeID     string
	PubKey     cryptotypes.PubKey
	Moniker    string
	APIAddress string
	RPCAddress string
	P2PAddress string
	Address    sdk.AccAddress
	ValAddress sdk.ValAddress
	RPCClient  cmtclient.Client

	app      servertypes.Application
	tmNode   *node.Node
	api      *api.Server
	grpc     *grpc.Server
	grpcWeb  *http.Server
	errGroup *errgroup.Group
	cancelFn context.CancelFunc
}

// ValidatorI expose a validator's context and configuration
type ValidatorI interface {
	GetCtx() *server.Context
	GetClientCtx() client.Context
	GetAppConfig() *srvconfig.Config
	GetAddress() sdk.AccAddress
	GetValAddress() sdk.ValAddress
}

var _ ValidatorI = Validator{}

func (v Validator) GetCtx() *server.Context {
	return v.Ctx
}

func (v Validator) GetAppConfig() *srvconfig.Config {
	return v.AppConfig
}

func (v Validator) GetClientCtx() client.Context {
	return v.ClientCtx
}

func (v Validator) GetAddress() sdk.AccAddress {
	return v.Address
}

func (v Validator) GetValAddress() sdk.ValAddress {
	return v.ValAddress
}
