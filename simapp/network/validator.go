package network

import (
	"context"
	"net/http"

	"github.com/cometbft/cometbft/node"
	cmtclient "github.com/cometbft/cometbft/rpc/client"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

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
	clientCtx  client.Context
	ctx        *server.Context
	dir        string
	nodeID     string
	pubKey     cryptotypes.PubKey
	moniker    string
	aPIAddress string
	rPCAddress string
	p2PAddress string
	address    sdk.AccAddress
	valAddress sdk.ValAddress
	rPCClient  cmtclient.Client

	app      servertypes.Application
	tmNode   *node.Node
	api      *api.Server
	grpc     *grpc.Server
	grpcWeb  *http.Server
	errGroup *errgroup.Group
	cancelFn context.CancelFunc
}

var _ ValidatorI = &Validator{}

func (v *Validator) GetCtx() *server.Context {
	return v.ctx
}

func (v *Validator) GetClientCtx() client.Context {
	return v.clientCtx
}

func (v *Validator) GetAppConfig() *srvconfig.Config {
	return v.AppConfig
}

func (v *Validator) GetAddress() sdk.AccAddress {
	return v.address
}

func (v *Validator) GetValAddress() sdk.ValAddress {
	return v.valAddress
}

func (v *Validator) GetAPIAddress() string {
	return v.aPIAddress
}

func (v *Validator) GetRPCAddress() string {
	return v.rPCAddress
}

func (v *Validator) GetPubKey() cryptotypes.PubKey {
	return v.pubKey
}

func (v *Validator) GetMoniker() string {
	return v.moniker
}
