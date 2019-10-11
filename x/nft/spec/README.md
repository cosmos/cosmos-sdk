# NFT Specification

## Overview

The NFT Module described here is meant to be used as a module across chains for managing non-fungible token that represent individual assets with unique features. This standard was first developed on Ethereum within the ERC-721 and the subsequent EIP of the same name. This standard utilized the features of the Ethereum blockchain as well as the restrictions. The subsequent ERC-1155 standard addressed some of the restrictions of Ethereum regarding storage costs and semi-fungible assets.

NFTs on application specific blockchains share some but not all features as their Ethereum brethren. Since application specific blockchains are more flexible in how their resources are utilized it makes sense that should have the option of exploiting those resources. This includes the aility to use strings as IDs and to optionally store metadata on chain. The user-flow of composability with smart contracts should also be rethought on application specific blockchains with regard to Inter-Blockchain Communication as it is a different design experience from communication between smart contracts.

## Contents

1. **[Concepts](./01_concepts.md)**
   - [NFT](./01_concepts.md#nft)
   - [Collections](./01_concepts.md#collections)
2. **[State](./02_state.md)**
   - [Collections](./02_state.md#collections)
   - [Owners](./02_state.md#owners)
3. **[Messages](./03_messages.md)**
   - [Transfer NFT](./03_messages.md#transfer-nft)
   - [Edit Metadata](./03_messages.md#edit-metadata)
   - [Mint NFT](./03_messages.md#mint-nft)
   - [Burn NFT](./03_messages.md#burn-nft)
4. **[Events](./04_events.md)**
5. **[Future Improvements](./05_future_improvements.md)**

## A Note on Metadata & IBC

The BaseNFT includes `tokenURI` in order to be backwards compatible with Ethereum based NFTs. However the `NFT` type is an interface that allows arbitrary metadata to be stored on chain should it need be. Originally the module included `name`, `description` and `image` to demonstrate these capabilities. They were removed in order for the NFT to be more efficient for use cases that don't include a need for that information to be stored on chain. A demonstration of including them will be included in a sample app. It is also under discussion to move all metadata to a separate module that can handle arbitrary amounts of data on chain and can be used to describe assets beyond Non-Fungible Tokens, like normal Fungible Tokens `Coin` that could describe attributes like decimal places and vesting status.

A stand-alone metadata Module would allow for independent standards to evolve regarding arbitrary asset types with expanding precision. The standards supported by [http://schema.org](http://schema.org) and the process of adding nested information is being considered as a starting point for that standard. The Blockchain Gaming Alliance is working on a metadata standard to be used for specifically blockchain gaming assets.

With regards to Inter-Blockchain Communication the responsibility of the integrity of the metadata should be left to the origin chain. If a secondary chain was responsible for storing the source of truth of the metadata for an asset tracking that source of truth would become difficult if not impossible to track. Since origin chains are where the design and use of the NFT is determined, it should be up to that origin chain to decide who can update metadata and under what circumstances. Secondary chains can use IBC queriers to check needed metadata or keep redundant copies of the metadata locally when they receive the NFT originally. In that case it should be up to te secondary chain to keep the metadata in sync if need be, similar to how layer 2 solutions keep metadata in sync with a source of truth using events.

## Custom App-Specific Handlers

Each message type comes with a default handler that can be used by default but will most likely be too limited for each use case. In order to make them useful for as many situations as possible, there are very few limitations on who can execute the Messages and do things like mint, burn or edit metadata. We recommend that custom handlers are created to add in custom logic and restrictions over when the Message types can be executed. Below is an example implementation for initializing the module within the module manager so that a custom handler can be added. This can be seen in the example [NFT app](https://github.com/okwme/cosmos-nft).

```go
// custom-handler.go

// OverrideNFTModule overrides the NFT module for custom handlers
type OverrideNFTModule struct {
  nft.AppModule
  k nft.Keeper
}

// NewHandler overwrites the legacy NewHandler in order to allow custom logic for handling the messages
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
        return HandleMsgMintNFTCustom(ctx, msg, k) // <-- This one is custom, the others fall back onto the default
      case types.MsgBurnNFT:
        return nft.HandleMsgBurnNFT(ctx, msg, k)
      default:
        errMsg := fmt.Sprintf("unrecognized nft message type: %T", msg)
        return sdk.ErrUnknownRequest(errMsg).Result()
    }
  }
}

// HandleMsgMintNFTCustom is a custom handler that handles MsgMintNFT
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

The default handlers are imported here with the NFT module and used for `MsgTransferNFT`, `MsgEditNFTMetadata` and `MsgBurnNFT`. The `MsgMintNFT` however is handled with a custom function called `HandleMsgMintNFTCustom`. This custom function also utilizes the imported NFT module handler `HandleMsgMintNFT`, but only after certain conditions are checked. In this case it checks a function called `checkTwilight` which returns a boolean. Only if `isTwilight` is true will the Message succeed.

This pattern of inheriting and utilizing the module handlers wrapped in custom logic should allow each application specific blockchain to use the NFT while customizing it to their specific requirements.
