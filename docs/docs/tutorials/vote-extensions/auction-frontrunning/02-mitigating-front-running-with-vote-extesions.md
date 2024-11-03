# Mitigating Front-running with Vote Extensions

## Table of Contents

- [Prerequisites](#prerequisites)
- [Implementing Structs for Vote Extensions](#implementing-structs-for-vote-extensions)
- [Implementing Handlers and Configuring Handlers](#implementing-handlers-and-configuring-handlers)

## Prerequisites

Before implementing vote extensions to mitigate front-running, ensure you have a module ready to implement the vote extensions with. If you need to create or reference a similar module, see `x/auction` for guidance.

In this section, we will discuss the steps to mitigate front-running using vote extensions. We will introduce new types within the `abci/types.go` file. These types will be used to handle the process of preparing proposals, processing proposals, and handling vote extensions.

### Implementing Structs for Vote Extensions

First, copy the following structs into the `abci/types.go` and each of these structs serves a specific purpose in the process of mitigating front-running using vote extensions:

```go
package abci

import (
 //import the necessary files
)

type PrepareProposalHandler struct {
 logger      log.Logger
 txConfig    client.TxConfig
 cdc         codec.Codec
 mempool     *mempool.ThresholdMempool
 txProvider  provider.TxProvider
 keyname     string
 runProvider bool
}
```

The `PrepareProposalHandler` struct is used to handle the preparation of a proposal in the consensus process. It contains several fields: logger for logging information and errors, txConfig for transaction configuration, cdc (Codec) for encoding and decoding transactions, mempool for referencing the set of unconfirmed transactions, txProvider for building the proposal with transactions, keyname for the name of the key used for signing transactions, and runProvider, a boolean flag indicating whether the provider should be run to build the proposal.

```go
type ProcessProposalHandler struct {
 TxConfig client.TxConfig
 Codec    codec.Codec
 Logger   log.Logger
}
```

After the proposal has been prepared and vote extensions have been included, the `ProcessProposalHandler` is used to process the proposal. This includes validating the proposal and the included vote extensions. The `ProcessProposalHandler` allows you to access the transaction configuration and codec, which are necessary for processing the vote extensions.

```go
type VoteExtHandler struct {
 logger       log.Logger
 currentBlock int64
 mempool      *mempool.ThresholdMempool
 cdc          codec.Codec
}
```

This struct is used to handle vote extensions. It contains a logger for logging events, the current block number, a mempool for storing transactions, and a codec for encoding and decoding. Vote extensions are a key part of the process to mitigate front-running, as they allow for additional information to be included with each vote.

```go
type InjectedVoteExt struct {
 VoteExtSigner []byte
 Bids          [][]byte
}

type InjectedVotes struct {
 Votes []InjectedVoteExt
}
```

These structs are used to handle injected vote extensions. They include the signer of the vote extension and the bids associated with the vote extension. Each byte array in Bids is a serialised form of a bid transaction. Injected vote extensions are used to add additional information to a vote after it has been created, which can be useful for adding context or additional data to a vote. The serialised bid transactions provide a way to include complex transaction data in a compact, efficient format.

```go
type AppVoteExtension struct {
 Height int64
 Bids   [][]byte
}
```

This struct is used for application vote extensions. It includes the height of the block and the bids associated with the vote extension. Application vote extensions are used to add additional information to a vote at the application level, which can be useful for adding context or additional data to a vote that is specific to the application.

```go
type SpecialTransaction struct {
 Height int
 Bids   [][]byte
}
```

This struct is used for special transactions. It includes the height of the block and the bids associated with the transaction. Special transactions are used for transactions that need to be handled differently from regular transactions, such as transactions that are part of the process to mitigate front-running.

### Implementing Handlers and Configuring Handlers

To establish the `VoteExtensionHandler`, follow these steps:

1. Navigate to the `abci/proposal.go` file. This is where we will implement the `VoteExtensionHandler``.

2. Implement the `NewVoteExtensionHandler` function. This function is a constructor for the `VoteExtHandler` struct. It takes a logger, a mempool, and a codec as parameters and returns a new instance of `VoteExtHandler`.

```go
func NewVoteExtensionHandler(lg log.Logger, mp *mempool.ThresholdMempool, cdc codec.Codec) *VoteExtHandler {  
   return &VoteExtHandler{  
      logger:  lg,  
      mempool: mp,  
      cdc:     cdc,  
   }  
}
```

3. Implement the `ExtendVoteHandler()` method. This method should handle the logic of extending votes, including inspecting the mempool and submitting a list of all pending bids. This will allow you to access the list of unconfirmed transactions in the abci.`RequestPrepareProposal` during the ensuing block.

```go
func (h *VoteExtHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
 return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
      h.logger.Info(fmt.Sprintf("Extending votes at block height : %v", req.Height))

 voteExtBids := [][]byte{}

 // Get mempool txs
 itr := h.mempool.SelectPending(context.Background(), nil)
 for itr != nil {
  tmptx := itr.Tx()
  sdkMsgs := tmptx.GetMsgs()

  // Iterate through msgs, check for any bids
  for _, msg := range sdkMsgs {
   switch msg := msg.(type) {
   case *nstypes.MsgBid:
   // Marshal sdk bids to []byte
    bz, err := h.cdc.Marshal(msg)
    if err != nil {
     h.logger.Error(fmt.Sprintf("Error marshalling VE Bid : %v", err))
     break
    }
    voteExtBids = append(voteExtBids, bz)
   default:
   }
  }

  // Move tx to ready pool
  err := h.mempool.Update(context.Background(), tmptx)
  
  // Remove tx from app side mempool
  if err != nil {
   h.logger.Info(fmt.Sprintf("Unable to update mempool tx: %v", err))
  }
  
  itr = itr.Next()
 }

 // Create vote extension
 voteExt := AppVoteExtension{
 Height: req.Height,
 Bids: voteExtBids,
 }

 // Encode Vote Extension
 bz, err := json.Marshal(voteExt)
  if err != nil {
  return nil, fmt.Errorf("Error marshalling VE: %w", err)
 }

 return &abci.ResponseExtendVote{VoteExtension: bz}, nil
}
```

4. Configure the handler in `app/app.go` as shown below

```go
bApp := baseapp.NewBaseApp(AppName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
voteExtHandler := abci2.NewVoteExtensionHandler(logger, mempool, appCodec)
bApp.SetExtendVoteHandler(voteExtHandler.ExtendVoteHandler())
```

To give a bit of context on what is happening above, we first create a new instance of `VoteExtensionHandler` with the necessary dependencies (logger, mempool, and codec). Then, we set this handler as the `ExtendVoteHandler` for our application. This means that whenever a vote needs to be extended, our custom `ExtendVoteHandler()` method will be called.

To test if vote extensions have been propagated, add the following to the `PrepareProposalHandler`:

```go
if req.Height > 2 {  
   voteExt := req.GetLocalLastCommit()  
   h.logger.Info(fmt.Sprintf("🛠️ :: Get vote extensions: %v", voteExt))  
}
```

This is how the whole function should look:

```go
func (h *PrepareProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
 return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
  h.logger.Info(fmt.Sprintf("🛠️ :: Prepare Proposal"))
  var proposalTxs [][]byte

  var txs []sdk.Tx

  // Get Vote Extensions
  if req.Height > 2 {
   voteExt := req.GetLocalLastCommit()
   h.logger.Info(fmt.Sprintf("🛠️ :: Get vote extensions: %v", voteExt))
  }

  itr := h.mempool.Select(context.Background(), nil)
  for itr != nil {
   tmptx := itr.Tx()

   txs = append(txs, tmptx)
   itr = itr.Next()
  }
  h.logger.Info(fmt.Sprintf("🛠️ :: Number of Transactions available from mempool: %v", len(txs)))

  if h.runProvider {
   tmpMsgs, err := h.txProvider.BuildProposal(ctx, txs)
   if err != nil {
    h.logger.Error(fmt.Sprintf("❌️ :: Error Building Custom Proposal: %v", err))
   }
   txs = tmpMsgs
  }

  for _, sdkTxs := range txs {
   txBytes, err := h.txConfig.TxEncoder()(sdkTxs)
   if err != nil {
    h.logger.Info(fmt.Sprintf("❌~Error encoding transaction: %v", err.Error()))
   }
   proposalTxs = append(proposalTxs, txBytes)
  }

  h.logger.Info(fmt.Sprintf("🛠️ :: Number of Transactions in proposal: %v", len(proposalTxs)))

  return &abci.ResponsePrepareProposal{Txs: proposalTxs}, nil
 }
}
```

As mentioned above, we check if vote extensions have been propagated, you can do this by checking the logs for any relevant messages such as `🛠️ :: Get vote extensions:`. If the logs do not provide enough information, you can also reinitialise your local testing environment by running the `./scripts/single_node/setup.sh` script again.

5. Implement the `ProcessProposalHandler()`. This function is responsible for processing the proposal. It should handle the logic of processing vote extensions, including inspecting the proposal and validating the bids.

```go
func (h *ProcessProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
 return func(ctx sdk.Context, req *abci.RequestProcessProposal) (resp *abci.ResponseProcessProposal, err error) {
  h.Logger.Info(fmt.Sprintf("⚙️ :: Process Proposal"))

  // The first transaction will always be the Special Transaction
  numTxs := len(req.Txs)

  h.Logger.Info(fmt.Sprintf("⚙️:: Number of transactions :: %v", numTxs))

  if numTxs >= 1 {
   var st SpecialTransaction
   err = json.Unmarshal(req.Txs[0], &st)
   if err != nil {
    h.Logger.Error(fmt.Sprintf("❌️:: Error unmarshalling special Tx in Process Proposal :: %v", err))
   }
   if len(st.Bids) > 0 {
    h.Logger.Info(fmt.Sprintf("⚙️:: There are bids in the Special Transaction"))
    var bids []nstypes.MsgBid
    for i, b := range st.Bids {
     var bid nstypes.MsgBid
     h.Codec.Unmarshal(b, &bid)
     h.Logger.Info(fmt.Sprintf("⚙️:: Special Transaction Bid No %v :: %v", i, bid))
     bids = append(bids, bid)
    }
    // Validate Bids in Tx
    txs := req.Txs[1:]
    ok, err := ValidateBids(h.TxConfig, bids, txs, h.Logger)
    if err != nil {
     h.Logger.Error(fmt.Sprintf("❌️:: Error validating bids in Process Proposal :: %v", err))
     return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
    }
    if !ok {
     h.Logger.Error(fmt.Sprintf("❌️:: Unable to validate bids in Process Proposal :: %v", err))
     return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
    }
    h.Logger.Info("⚙️:: Successfully validated bids in Process Proposal")
   }
  }

  return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
 }
}
```

6. Implement the `ProcessVoteExtensions()` function. This function should handle the logic of processing vote extensions, including validating the bids.

```go
func processVoteExtensions(req *abci.RequestPrepareProposal, log log.Logger) (SpecialTransaction, error) {
 log.Info(fmt.Sprintf("🛠️ :: Process Vote Extensions"))

 // Create empty response
 st := SpecialTransaction{
  0,
  [][]byte{},
 }

 // Get Vote Ext for H-1 from Req
 voteExt := req.GetLocalLastCommit()
 votes := voteExt.Votes

 // Iterate through votes
 var ve AppVoteExtension
 for _, vote := range votes {
  // Unmarshal to AppExt
  err := json.Unmarshal(vote.VoteExtension, &ve)
  if err != nil {
   log.Error(fmt.Sprintf("❌ :: Error unmarshalling Vote Extension"))
  }

  st.Height = int(ve.Height)

  // If Bids in VE, append to Special Transaction
  if len(ve.Bids) > 0 {
   log.Info("🛠️ :: Bids in VE")
   for _, b := range ve.Bids {
    st.Bids = append(st.Bids, b)
   }
  }
 }

 return st, nil
}
```

7. Configure the `ProcessProposalHandler()` in app/app.go:

```go
processPropHandler := abci2.ProcessProposalHandler{app.txConfig, appCodec, logger}
bApp.SetProcessProposal(processPropHandler.ProcessProposalHandler())
```

This sets the `ProcessProposalHandler()` for our application. This means that whenever a proposal needs to be processed, our custom `ProcessProposalHandler()` method will be called.

To test if the proposal processing and vote extensions are working correctly, you can check the logs for any relevant messages. If the logs do not provide enough information, you can also reinitialize your local testing environment by running `./scripts/single_node/setup.sh` script.
