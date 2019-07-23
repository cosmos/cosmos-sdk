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

The BaseNFT includes what was considered part of the off-chain metadata for the original ERC-721. These were the fields expected in the JSON object that was expected to be returned by resolvingn the URI found in the on-chain field `tokenURI`. You can see that `tokenURI` is also included here. This is to represent that it is possible to store this data on chain while allowing more data to be stored off chain. While this was the format chosen for the first version of the Cosmos NFT, it is under discussion to move all metadata to a separate module that can handle arbitrary amounts of data on chain and can be used to describe assets beyond Non-Fungible Tokens.

A stand-alone metadata Module would allow for independent standards to evolve regarding arbitrary asset types with expanding precision. The standards supported by [http://schema.org](http://schema.org) and the process of adding nested information is being considered as a starting point for that standard.

With regards to Inter-Blockchain Communication the responsibility of the integrity of the metadata should be left to the origin chain. If a secondary chain was responsible for storing the source of truth of the metadata for an asset tracking that source of truth would become difficult if not impossible. Since origin chains are where the design and use of the NFT is determined, it should be up to that origin chain to decide who can update metadata and under what circumstances. Secondary chains can use IBC queriers to check needed metadata or keep redundant copies of the metadata locally. In that case it should be up to te secondary chain to keep the metadata in sync, similar to how layer 2 solutions keep metadata in sync with a source of truth using events.

## Custom App-Specific Handlers

Each message type comes with a default handler that can be used by default but will most likely be too limited for each use case. We recommend that custom handlers are created to add in custom logic and restrictions over when the Message types can be executed. Below is a recomended method for initializing the module within the module manager so that a custom handler can be added. This can be seen in the example [NFT app](https://github.com/okwme/cosmos-nft).

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