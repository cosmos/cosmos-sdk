---
sidebar_position: 1
---

# Transaction Lifecycle

:::note Synopsis
This document describes the lifecycle of a transaction from creation to committed state changes. Transaction definition is described in a [different doc](../advanced/01-transactions.md). The transaction is referred to as `Tx`.
:::

:::note Pre-requisite Readings

* [Anatomy of a Cosmos SDK Application](./00-app-anatomy.md)
:::


## Transaction Creation

One of the main application interfaces is the command-line interface. The transaction `Tx` can be created by the user inputting a command in the following format from the [command-line](../advanced/07-cli.md), providing the type of transaction in `[command]`, arguments in `[args]`, and configurations such as gas prices in `[flags]`:

```bash
[appname] tx [command] [args] [flags]
```

This command automatically **creates** the transaction, **signs** it using the account's private key, and **broadcasts** it to the specified peer node.

There are several required and optional flags for transaction creation. The `--from` flag specifies which [account](./03-accounts.md) the transaction is originating from. For example, if the transaction is sending coins, the funds are drawn from the specified `from` address.

### Gas and Fees

Additionally, there are several [flags](../advanced/07-cli.md) users can use to indicate how much they are willing to pay in [fees](./04-gas-fees.md):

* `--gas` refers to how much [gas](./04-gas-fees.md), which represents computational resources, `Tx` consumes. Gas is dependent on the transaction and is not precisely calculated until execution, but can be estimated by providing `auto` as the value for `--gas`.
* `--gas-adjustment` (optional) can be used to scale `gas` up in order to avoid underestimating. For example, users can specify their gas adjustment as 1.5 to use 1.5 times the estimated gas.
* `--gas-prices` specifies how much the user is willing to pay per unit of gas, which can be one or multiple denominations of tokens. For example, `--gas-prices=0.025uatom, 0.025upho` means the user is willing to pay 0.025uatom AND 0.025upho per unit of gas.
* `--fees` specifies how much in fees the user is willing to pay in total.
* `--timeout-height` specifies a block timeout height to prevent the tx from being committed past a certain height.

The ultimate value of the fees paid is equal to the gas multiplied by the gas prices. In other words, `fees = ceil(gas * gasPrices)`. Thus, since fees can be calculated using gas prices and vice versa, the users specify only one of the two.

Later, validators decide whether or not to include the transaction in their block by comparing the given or calculated `gas-prices` to their local `min-gas-prices`. `Tx` is rejected if its `gas-prices` is not high enough, so users are incentivized to pay more.

### CLI Example

Users of the application `app` can enter the following command into their CLI to generate a transaction to send 1000uatom from a `senderAddress` to a `recipientAddress`. The command specifies how much gas they are willing to pay: an automatic estimate scaled up by 1.5 times, with a gas price of 0.025uatom per unit gas.

```bash
appd tx send <recipientAddress> 1000uatom --from <senderAddress> --gas auto --gas-adjustment 1.5 --gas-prices 0.025uatom
```

### Other Transaction Creation Methods

The command-line is an easy way to interact with an application, but `Tx` can also be created using a [gRPC or REST interface](../advanced/06-grpc_rest.md) or some other entry point defined by the application developer. From the user's perspective, the interaction depends on the web interface or wallet they are using (e.g. creating `Tx` using [Keplr](https://www.keplr.app/) and signing it with a Ledger Nano S).

## Transaction Broadcasting

This is the next phase, where a transactison is sent from a client (such as a wallet or a command-line interface) to the network of nodes. This process is consensus-agnostic, meaning it can work with various consensus engines.

Below are the steps involved in transaction broadcasting:

1. **Transaction Creation and Signing:**
Transactions are created and signed using the client's private key to ensure authenticity and integrity.

2. **Broadcasting to the Network:**
The signed transaction is sent to the network. This is handled by the `BroadcastTx` function in the client context.

3. **Network Propagation:**
Once received by a node, the transaction is propagated to other nodes in the network. This ensures that all nodes have a copy of the transaction.

4. **Consensus Engine Interaction:**
The specific method of broadcasting may vary depending on the consensus engine used. The SDK's design allows for easy integration with any consensus engine by configuring the clientCtx appropriately.

The function `BroadcastTx` in `client/tx/tx.go` demonstrates how a transaction is prepared, signed, and broadcasted. Here's the relevant part of the function that handles the broadcasting:

```go
res, err := clientCtx.BroadcastTx(txBytes)
if err != nil {
 return err
}
```

**Configuration:**

To adapt this function for different consensus engines, ensure that the `clientCtx` is configured with the correct network settings and transaction handling mechanisms for your chosen engine. This might involve setting up specific encoders, decoders, and network endpoints that are compatible with the engine.

## Transaction Processing 

After a transaction is broadcasted to the network, it undergoes several processing steps to ensure its validity. These steps are run by the application's `BaseApp`, which is in chargee of trasaction processing. Within the SDK we wrap the `BaseApp` with a runtime layer defined in `runtime/app.go`. This is where `BaseApp` is extended with additional functionality to handle transactions allowing for more flexibility and customisation, for example being able to swap out CometBFT for another consensus engine. The key steps in transaction processing are:

### Decoding

The transaction is decoded from its binary format into a structured format that the application can understand.

### Validation

Preliminary checks are performed. These include signature verification to ensure the transaction hasn't been tampered with and checking if the transaction meets the minimum fee requirements, which is handled by the AnteHandler. The Antehandler is invoked during the `runTx` method in `BaseApp`.

### Routing

How Routing Works:
The transaction is routed to the appropriate module based on the message type. Each message type is associated with a specific module, which is responsible for processing the message. The `BaseApp` uses a `MsgServiceRouter` to direct the transaction to the correct module.

1. **Transaction Type Identification:** 
Each transaction contains one or more messages (`sdk.Msg`), and each message has a `Type()` method that identifies its type. This type is used to determine the appropriate module to handle the message.
2. **Module Routing:** 
The BaseApp holds a `MsgServiceRouter` which maps each message type to a specific module's handler. When a transaction is processed, `BaseApp` uses this router to direct the message to the correct module.


### Example of Routing

Let's say there is a transaction that involves transferring tokens. The message type might be `MsgSend`, and the `MsgServiceRouter` in `BaseApp` would route this message to the bank module's handler. The bank module would then validate the transaction details (like sender balance) and update the state to reflect the transfer if valid...

### Module Execution

1. The transaction is first routed to the appropriate module based on the message type. This is handled by the `MsgServiceRouter`.

2. Each module has specific handlers that are triggered once the message is routed to the module. These handlers contain the logic needed to process the transaction, such as updating account balances, transferring tokens, or other state changes.

3. During the execution, the module's handler will modify the state as required by the business logic. This could involve writing to the module's portion of the state store.

4. Modules can emit events and log information during execution, which are used for monitoring and querying transaction outcomes.

```go 
func handleMsgSend(ctx sdk.Context, keeper BankKeeper, msg MsgSend) error {
    if keeper.GetBalance(ctx, msg.Sender).Amount < msg.Amount {
        return sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, "sender does not have enough tokens")
    }
    keeper.SendCoins(ctx, msg.Sender, msg.Receiver, msg.Amount)
    ctx.EventManager().EmitEvent(
        sdk.NewEvent("transfer", sdk.NewAttribute("from", msg.Sender.String()), sdk.NewAttribute("to", msg.Receiver.String()), sdk.NewAttribute("amount", msg.Amount.String())),
    )
    return nil
}
```

## Post-Transaction Handling

After execution, any additional actions that need to be taken are processed. This could include updating logs, sending events, or handling errors.

These steps are managed by `BaseApp` in the Cosmos SDK, which routes transactions to the appropriate handlers and manages state transitions.

After a transaction is executed in the Cosmos SDK, several steps are taken to finalise the process:

1. Event Emission: Modules emit events that can be used for logging, monitoring, or triggering other workflows. These events are collected during the transaction execution.

2. Logging: Information about the transaction execution, such as success or failure, and any significant state changes, are logged for audit and diagnostic purposes.

3. Error Handling: If any errors occur during transaction execution, they are handled appropriately, which may include rolling back certain operations to maintain state consistency.

4. State Commitment: Changes made to the state during the transaction are finalized and written to the blockchain. This step is crucial as it ensures that all state transitions are permanently recorded.

After post-transaction handling in the Cosmos SDK, the exact sequence of the transaction lifecycle is dependent on the consensus mechanism used. This includes how transactions are grouped into blocks, how blocks are validated, and how consensus is achieved among validators to commit the block to the blockchain. Each consensus protocol may implement these steps differently to ensure network agreement and maintain the integrity of the blockchain state.
