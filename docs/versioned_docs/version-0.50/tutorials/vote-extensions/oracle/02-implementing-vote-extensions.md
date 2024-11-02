# Implementing Vote Extensions

## Implement ExtendVote

First we’ll create the `OracleVoteExtension` struct, this is the object that will be marshaled as bytes and signed by the validator.

In our example we’ll use JSON to marshal the vote extension for simplicity but we recommend to find an encoding that produces a smaller output, given that large vote extensions could impact CometBFT’s performance. Custom encodings and compressed bytes can be used out of the box.

```go
// OracleVoteExtension defines the canonical vote extension structure.
type OracleVoteExtension struct {
 Height int64
 Prices map[string]math.LegacyDec
}
```

Then we’ll create a `VoteExtensionsHandler` struct that contains everything we need to query for prices.

```go
type VoteExtHandler struct {
 logger          log.Logger
 currentBlock    int64                            // current block height
 lastPriceSyncTS time.Time                        // last time we synced prices
 providerTimeout time.Duration                    // timeout for fetching prices from providers
 providers       map[string]Provider              // mapping of provider name to provider (e.g. Binance -> BinanceProvider)
 providerPairs   map[string][]keeper.CurrencyPair // mapping of provider name to supported pairs (e.g. Binance -> [ATOM/USD])

 Keeper keeper.Keeper // keeper of our oracle module
}
```

Finally, a function that returns `sdk.ExtendVoteHandler` is needed too, and this is where our vote extension logic will live.

```go
func (h *VoteExtHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
    return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
        // here we'd have a helper function that gets all the prices and does a weighted average using the volume of each market
        prices := h.getAllVolumeWeightedPrices()

        voteExt := OracleVoteExtension{
            Height: req.Height,
            Prices: prices,
        }
        
        bz, err := json.Marshal(voteExt)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal vote extension: %w", err)
        }

        return &abci.ResponseExtendVote{VoteExtension: bz}, nil
    }
}
```

As you can see above, the creation of a vote extension is pretty simple and we just have to return bytes. CometBFT will handle the signing of these bytes for us. We ignored the process of getting the prices but you can see a more complete example [here:](https://github.com/cosmos/sdk-tutorials/blob/master/tutorials/oracle/base/x/oracle/abci/vote_extensions.go)

Here we’ll do some simple checks like:

* Is the vote extension unmarshaled correctly?
* Is the vote extension for the right height?
* Some other validation, for example, are the prices from this extension too deviated from my own prices? Or maybe checks that can detect malicious behavior.

```go
func (h *VoteExtHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
    return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
        var voteExt OracleVoteExtension
        err := json.Unmarshal(req.VoteExtension, &voteExt)
        if err != nil {
            return nil, fmt.Errorf("failed to unmarshal vote extension: %w", err)
        }
        
        if voteExt.Height != req.Height {
            return nil, fmt.Errorf("vote extension height does not match request height; expected: %d, got: %d", req.Height, voteExt.Height)
        }

        // Verify incoming prices from a validator are valid. Note, verification during
        // VerifyVoteExtensionHandler MUST be deterministic. For brevity and demo
        // purposes, we omit implementation.
        if err := h.verifyOraclePrices(ctx, voteExt.Prices); err != nil {
            return nil, fmt.Errorf("failed to verify oracle prices from validator %X: %w", req.ValidatorAddress, err)
        }

        return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
    }
}
```

## Implement PrepareProposal

```go
type ProposalHandler struct {
    logger   log.Logger
    keeper   keeper.Keeper // our oracle module keeper
    valStore baseapp.ValidatorStore // to get the current validators' pubkeys
}
```

And we create the struct for our “special tx”, that will contain the prices and the votes so validators can later re-check in ProcessPRoposal that they get the same result than the block’s proposer. With this we could also check if all the votes have been used by comparing the votes received in ProcessProposal.

```go
type StakeWeightedPrices struct {
    StakeWeightedPrices map[string]math.LegacyDec
    ExtendedCommitInfo  abci.ExtendedCommitInfo
}
```

Now we create the `PrepareProposalHandler`. In this step we’ll first check if the vote extensions’ signatures are correct using a helper function called ValidateVoteExtensions from the baseapp package.

```go
func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
    return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
        err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), req.LocalLastCommit)
        if err != nil {
            return nil, err
        }
...
```

Then we proceed to make the calculations only if the current height if higher than the height at which vote extensions have been enabled. Remember that vote extensions are made available to the block proposer on the next block at which they are produced/enabled.

```go
...
        proposalTxs := req.Txs

        if req.Height > ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
            stakeWeightedPrices, err := h.computeStakeWeightedOraclePrices(ctx, req.LocalLastCommit)
            if err != nil {
                return nil, errors.New("failed to compute stake-weighted oracle prices")
            }

            injectedVoteExtTx := StakeWeightedPrices{
                StakeWeightedPrices: stakeWeightedPrices,
                ExtendedCommitInfo:  req.LocalLastCommit,
            }
...
```

Finally we inject the result as a transaction at a specific location, usually at the beginning of the block:

## Implement ProcessProposal

Now we can implement the method that all validators will execute to ensure the proposer is doing his work correctly.

Here, if vote extensions are enabled, we’ll check if the tx at index 0 is an injected vote extension

```go
func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
    return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
        if req.Height > ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
            var injectedVoteExtTx StakeWeightedPrices
            if err := json.Unmarshal(req.Txs[0], &injectedVoteExtTx); err != nil {
                h.logger.Error("failed to decode injected vote extension tx", "err", err)
                return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
            }
...
```

Then we re-validate the vote extensions signatures using
baseapp.ValidateVoteExtensions, re-calculate the results (just like in PrepareProposal) and compare them with the results we got from the injected tx.

```go
            err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), injectedVoteExtTx.ExtendedCommitInfo)
            if err != nil {
                return nil, err
            }

            // Verify the proposer's stake-weighted oracle prices by computing the same
            // calculation and comparing the results. We omit verification for brevity
            // and demo purposes.
            stakeWeightedPrices, err := h.computeStakeWeightedOraclePrices(ctx, injectedVoteExtTx.ExtendedCommitInfo)
            if err != nil {
                return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
            }
            
            if err := compareOraclePrices(injectedVoteExtTx.StakeWeightedPrices, stakeWeightedPrices); err != nil {
                return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
            }
        }

        return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
    }
}
```

Important: In this example we avoided using the mempool and other basics, please refer to the DefaultProposalHandler for a complete implementation: [https://github.com/cosmos/cosmos-sdk/blob/v0.50.1/baseapp/abci_utils.go](https://github.com/cosmos/cosmos-sdk/blob/v0.50.1/baseapp/abci_utils.go)

## Implement PreBlocker

Now validators are extending their vote, verifying other votes and including the result in the block. But how do we actually make use of this result? This is done in the PreBlocker which is code that is run before any other code during FinalizeBlock so we make sure we make this information available to the chain and its modules during the entire block execution (from BeginBlock).

At this step we know that the injected tx is well-formatted and has been verified by the validators participating in consensus, so making use of it is straightforward. Just check if vote extensions are enabled, pick up the first transaction and use a method in your module’s keeper to set the result.

```go
func (h *ProposalHandler) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
    res := &sdk.ResponsePreBlock{}
    if len(req.Txs) == 0 {
        return res, nil
    }

    if req.Height > ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
        var injectedVoteExtTx StakeWeightedPrices
        if err := json.Unmarshal(req.Txs[0], &injectedVoteExtTx); err != nil {
            h.logger.Error("failed to decode injected vote extension tx", "err", err)
            return nil, err
        }

        // set oracle prices using the passed in context, which will make these prices available in the current block
        if err := h.keeper.SetOraclePrices(ctx, injectedVoteExtTx.StakeWeightedPrices); err != nil {
            return nil, err
        }
    }
    return res, nil
}

```

## Conclusion

In this tutorial, we've created a simple price oracle module that incorporates vote extensions. We've seen how to implement `ExtendVote`, `VerifyVoteExtension`, `PrepareProposal`, `ProcessProposal`, and `PreBlocker` to handle the voting and verification process of vote extensions, as well as how to make use of the results during the block execution.
