# Concepts

## NFT

The `NFT` Interface inherits the BaseNFT struct and includes getter functions for the asset data. It also lincludes a Stringer function in order to print the struct. The interface may change if metadata is moved to it’s own module as it might no longer be necessary for the flexibility of an interface.

```go
// NFT non fungible token interface
type NFT interface {
  GetID() string
  GetOwner() sdk.AccAddress
  SetOwner(address sdk.AccAddress)
  GetName() string
  GetDescription() string
  GetImage() string
  GetTokenURI() string
  EditMetadata(name, description, image, tokenURI string)
  String() string
}
```

## Collections

A Collection is used to organized sets of NFTs. It contains the denomination of the NFT instead of storing it within each NFT. This saves storage space by removing redundancy.

```go
// Collection of non fungible tokens
type Collection struct {
  Denom string `json:"denom,omitempty"` // name of the collection; not exported to clients
  NFTs  NFTs   `json:"nfts"`            // NFTs that belong to a collection
}
```

## Owner

An Owner is a struct that includes information about all NFTs owned by a single account. It would be possible to retrieve this information by looping through all Collections but that process could become computationaly prohibitive so a more efficient retrieval system is to store redundant information limited to the token ID by owner.

```go
// Owner of non fungible tokens
type Owner struct {
  Address       sdk.AccAddress `json:"address"`
  IDCollections IDCollections  `json:"IDCollections"`
}
```

## Custom App-Specific Handlers

Each Message type comes with a default handler that can be used by default but will most likely be too limited for each use case. We recommend that custom handlers are created to add in custom logic and restrictions over when the Message types can be executed. Below is a recomended method for initializing the module within the module manager so that a custom handler can be added. This can be seen in the example [NFT app](https://github.com/okwme/cosmos-nft).

```go
// custom-handler.go

// OverrideNFTModule overrides the NFT module for custom handlers
type OverrideNFTModule struct {
  nft.AppModule
  k nft.Keeper
}

// NewHandler module handler for the OerrideNFTModule
func (am OverrideNFTModule) NewHandler() sdk.Handler {
  return CustomNFTHandler(am.k)
}

// NewOverrideNFTModule generates a new NFT Module
func NewOverrideNFTModule(appModule nft.AppModule, keeper nft.Keeper) OverrideNFTModule {
  return OverrideNFTModule{
    AppModule: appModule,
    k:         keeper,
  }
}
```

You can see here that `OverrideNFTModule` is the same as `nft.AppModule` except for the `NewHandler()` method. This method now returns a new Handler called `CustomNFTHandler`. This custom handler can be seen below:

```go
// CustomNFTHandler routes the messages to the handlers
func CustomNFTHandler(k keeper.Keeper) sdk.Handler {
  return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
    switch msg := msg.(type) {
      case types.MsgTransferNFT:
        return nft.HandleMsgTransferNFT(ctx, msg, k)
      case types.MsgEditNFTMetadata:
        return nft.HandleMsgEditNFTMetadata(ctx, msg, k)
     case types.MsgMintNFT:
        return HandleMsgMintNFTCustom(ctx, msg, k)
      case types.MsgBurnNFT:
        return nft.HandleMsgBurnNFT(ctx, msg, k)
      default:
        errMsg := fmt.Sprintf("unrecognized nft message type: %T", msg)
        return sdk.ErrUnknownRequest(errMsg).Result()
    }
  }
}

// HandleMsgMintNFTCustom handles MsgMintNFT
func HandleMsgMintNFTCustom(ctx sdk.Context, msg types.MsgMintNFT, k keeper.Keeper,
) sdk.Result {

  isTwilight := checkTwilight(ctx)

  if isTwilight {
    return nft.HandleMsgMintNFT(ctx, msg, k)
  }

  errMsg := fmt.Sprintf("Can't mint astral bodies outside of twilight!")
  return sdk.ErrUnknownRequest(errMsg).Result()
  }
```

The default handlers are imported here with the NFT module and used for `MsgTransferNFT`, `MsgEditNFTMetadata` and `MsgBurnNFT`. The `MsgMintNFT` however is handled with a custom function called `HandleMsgMintNFTCustom`. This custom function also utilizes the imported NFT module handler `HandleMsgBurnNFT`, but only after certain conditions are checked. In this case it checks a function called `checkTwilight` which returns a boolean. Only if `isTwilight` is true will the Message succeed.

This pattern of inheriting and utlizing the module handlers wrapped in custom logic should allow each application specific blockchain to use the NFT while customizing it to their specific requirements.
