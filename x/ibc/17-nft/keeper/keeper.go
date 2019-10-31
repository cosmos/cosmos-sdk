package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/17-nft/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// DefaultPacketTimeout is the default packet timeout relative to the current block height
const (
	DefaultPacketTimeout = 1000 // NOTE: in blocks
)

// Keeper defines the IBC transfer keeper
type Keeper struct {
	storeKey  sdk.StoreKey
	cdc       *codec.Codec
	codespace sdk.CodespaceType
	prefix    []byte // prefix bytes for accessing the store

	clientKeeper     types.ClientKeeper
	connectionKeeper types.ConnectionKeeper
	channelKeeper    types.ChannelKeeper
	nftKeeper        types.NFTKeeper
}

// NewKeeper creates a new IBC transfer Keeper instance
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, codespace sdk.CodespaceType,
	clientKeeper types.ClientKeeper, connnectionKeeper types.ConnectionKeeper,
	channelKeeper types.ChannelKeeper, nftKeeper types.NFTKeeper,
) Keeper {

	return Keeper{
		storeKey:         key,
		cdc:              cdc,
		codespace:        sdk.CodespaceType(fmt.Sprintf("%s/%s", codespace, types.DefaultCodespace)), // "ibc/nft_transfer",
		prefix:           []byte(types.SubModuleName + "/"),                                          // "nft_transfer/"
		clientKeeper:     clientKeeper,
		connectionKeeper: connnectionKeeper,
		channelKeeper:    channelKeeper,
		nftKeeper:        nftKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s/%s", ibctypes.ModuleName, types.SubModuleName))
}
