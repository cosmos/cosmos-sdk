---
sidebar_position: 1
---

# Transactions

:::note Synopsis
`Transactions` are objects created by end-users to trigger state changes in the application.
:::

:::note Pre-requisite Readings

* [Anatomy of a Cosmos SDK Application](../beginner/00-app-anatomy.md)

:::

## Transactions

Transactions are comprised of metadata held in [contexts](./17-context.md) and [`sdk.Msg`s](../../build/building-modules/02-messages-and-queries.md) that trigger state changes within a module through the module's Protobuf [`Msg` service](../../build/building-modules/03-msg-services.md).

When users want to interact with an application and make state changes (e.g. sending coins), they create transactions. Each of a transaction's `sdk.Msg` must be signed using the private key associated with the appropriate account(s), before the transaction is broadcasted to the network. A transaction must then be included in a block, validated, and approved by the network through the consensus process. To read more about the lifecycle of a transaction, click [here](../beginner/01-tx-lifecycle.md).

## Type Definition

Transaction objects are Cosmos SDK types that implement the `Tx` interface

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/types/tx_msg.go#L53-L66
```

It contains the following methods:

* **GetMsgs:** unwraps the transaction and returns a list of contained `sdk.Msg`s - one transaction may have one or multiple messages, which are defined by module developers.
* **ValidateBasic:** lightweight, [_stateless_](../beginner/01-tx-lifecycle.md#types-of-checks) checks used by ABCI messages [`CheckTx`](./00-baseapp.md#checktx) and [`RunTx`](./00-baseapp.md#runtx) to make sure transactions are not invalid. For example, the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/main/x/auth) module's `ValidateBasic` function checks that its transactions are signed by the correct number of signers and that the fees do not exceed the user's maximum. When [`runTx`](./00-baseapp.md#runtx) is checking a transaction created from the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/main/x/auth/) module, it first runs `ValidateBasic` on each message, then runs the `auth` module AnteHandler which calls `ValidateBasic` for the transaction itself.
* **Hash()**: returns the unique identifier for the Tx.
* **GetMessages:** returns the list of `sdk.Msg`s contained in the transaction.
* **GetSenders:** returns the addresses of the signers who signed the transaction.
* **GetGasLimit:** returns the gas limit for the transaction. Returns `math.MaxUint64` for transactions with unlimited gas.
* **Bytes:** returns the encoded bytes of the transaction. This is typically cached after the first decoding of the transaction.

:::note
This function is different from the deprecated `sdk.Msg` [`ValidateBasic`](../beginner/01-tx-lifecycle.md#ValidateBasic) methods, which was performing basic validity checks on messages only.
:::

As a developer, you should rarely manipulate `Tx` directly, as `Tx` is really an intermediate type used for transaction generation. Instead, developers should prefer the `TxBuilder` interface, which you can learn more about [below](#transaction-generation).

### Signing Transactions

Every message in a transaction must be signed by the addresses specified by its `GetSigners`. The Cosmos SDK currently allows signing transactions in two different ways.

#### `SIGN_MODE_DIRECT` (preferred)

The most used implementation of the `Tx` interface is the Protobuf `Tx` message, which is used in `SIGN_MODE_DIRECT`:

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/proto/cosmos/tx/v1beta1/tx.proto#L15-L28
```

Because Protobuf serialization is not deterministic, the Cosmos SDK uses an additional `TxRaw` type to denote the pinned bytes over which a transaction is signed. Any user can generate a valid `body` and `auth_info` for a transaction, and serialize these two messages using Protobuf. `TxRaw` then pins the user's exact binary representation of `body` and `auth_info`, called respectively `body_bytes` and `auth_info_bytes`. The document that is signed by all signers of the transaction is `SignDoc` (deterministically serialized using [ADR-027](../../architecture/adr-027-deterministic-protobuf-serialization.md)):

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/proto/cosmos/tx/v1beta1/tx.proto#L50-L67
```

Once signed by all signers, the `body_bytes`, `auth_info_bytes` and `signatures` are gathered into `TxRaw`, whose serialized bytes are broadcasted over the network.

#### `SIGN_MODE_LEGACY_AMINO_JSON`

The legacy implementation of the `Tx` interface is the `StdTx` struct from `x/auth`:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/x/auth/migrations/legacytx/stdtx.go#L81-L91
```

The document signed by all signers is `StdSignDoc`:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/x/auth/migrations/legacytx/stdsign.go#L32-L45
```

which is encoded into bytes using Amino JSON. Once all signatures are gathered into `StdTx`, `StdTx` is serialized using Amino JSON, and these bytes are broadcasted over the network.

#### Other Sign Modes

The Cosmos SDK also provides a couple of other sign modes for particular use cases.

#### `SIGN_MODE_DIRECT_AUX`

`SIGN_MODE_DIRECT_AUX` is a sign mode released in the Cosmos SDK v0.46 which targets transactions with multiple signers. Whereas `SIGN_MODE_DIRECT` expects each signer to sign over both `TxBody` and `AuthInfo` (which includes all other signers' signer infos, i.e. their account sequence, public key and mode info), `SIGN_MODE_DIRECT_AUX` allows N-1 signers to only sign over `TxBody` and _their own_ signer info. Moreover, each auxiliary signer (i.e. a signer using `SIGN_MODE_DIRECT_AUX`) doesn't
need to sign over the fees:

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/proto/cosmos/tx/v1beta1/tx.proto#L69-L93
```

The use case is a multi-signer transaction, where one of the signers is appointed to gather all signatures, broadcast the signature and pay for fees, and the others only care about the transaction body. This generally allows for a better multi-signing UX. If Alice, Bob and Charlie are part of a 3-signer transaction, then Alice and Bob can both use `SIGN_MODE_DIRECT_AUX` to sign over the `TxBody` and their own signer info (no need an additional step to gather other signers' ones, like in `SIGN_MODE_DIRECT`), without specifying a fee in their SignDoc. Charlie can then gather both signatures from Alice and Bob, and
create the final transaction by appending a fee. Note that the fee payer of the transaction (in our case Charlie) must sign over the fees, so must use `SIGN_MODE_DIRECT` or `SIGN_MODE_LEGACY_AMINO_JSON`.


#### `SIGN_MODE_TEXTUAL`

`SIGN_MODE_TEXTUAL` is a new sign mode for delivering a better signing experience on hardware wallets and it is included in the v0.50 release. In this mode, the signer signs over the human-readable string representation of the transaction (CBOR) and makes all data being displayed easier to read. The data is formatted as screens, and each screen is meant to be displayed in its entirety even on small devices like the Ledger Nano.

There are also _expert_ screens, which will only be displayed if the user has chosen that option in its hardware device. These screens contain things like account number, account sequence and the sign data hash.

Data is formatted using a set of `ValueRenderer` which the SDK provides defaults for all the known messages and value types. Chain developers can also opt to implement their own `ValueRenderer` for a type/message if they'd like to display information differently.

If you wish to learn more, please refer to [ADR-050](../../architecture/adr-050-sign-mode-textual.md).

#### Custom Sign modes

There is the opportunity to add your own custom sign mode to the Cosmos-SDK.  While we can not accept the implementation of the sign mode to the repository, we can accept a pull request to add the custom signmode to the SignMode enum located [here](https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/proto/cosmos/tx/signing/v1beta1/signing.proto#L9-L17)

## Transaction Process

The process of an end-user sending a transaction is:

* decide on the messages to put into the transaction,
* generate the transaction using the Cosmos SDK's `TxBuilder`,
* broadcast the transaction using one of the available interfaces.

The next paragraphs will describe each of these components, in this order.

### Messages

:::tip
Module `sdk.Msg`s are not to be confused with [ABCI Messages](https://docs.cometbft.com/v1.0/spec/abci/) which define interactions between the CometBFT and application layers.
:::

**Messages** (or `sdk.Msg`s) are module-specific objects that trigger state transitions within the scope of the module they belong to. Module developers define the messages for their module by adding methods to the Protobuf [`Msg` service](../../build/building-modules/03-msg-services.md), and also implement the corresponding `MsgServer`.

Each `sdk.Msg`s is related to exactly one Protobuf [`Msg` service](../../build/building-modules/03-msg-services.md) RPC, defined inside each module's `tx.proto` file. A SDK app router automatically maps every `sdk.Msg` to a corresponding RPC. Protobuf generates a `MsgServer` interface for each module `Msg` service, and the module developer needs to implement this interface.
This design puts more responsibility on module developers, allowing application developers to reuse common functionalities without having to implement state transition logic repetitively.

To learn more about Protobuf `Msg` services and how to implement `MsgServer`, click [here](../../build/building-modules/03-msg-services.md).

While messages contain the information for state transition logic, a transaction's other metadata and relevant information are stored in the `TxBuilder` and `Context`.

### Transaction Generation

The `TxBuilder` interface contains data closely related with the generation of transactions, which an end-user can freely set to generate the desired transaction:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/client/tx_config.go#L39-L57
```

* `Msg`s, the array of [messages](#messages) included in the transaction.
* `GasLimit`, option chosen by the users for how to calculate how much gas they will need to pay.
* `Memo`, a note or comment to send with the transaction.
* `FeeAmount`, the maximum amount the user is willing to pay in fees.
* `TimeoutHeight`, block height until which the transaction is valid.
* `Signatures`, the array of signatures from all signers of the transaction.

As there are currently two sign modes for signing transactions, there are also two implementations of `TxBuilder`:

* [builder](https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/x/auth/tx/builder.go#L79-L98) for creating transactions for `SIGN_MODE_DIRECT`,
* [StdTxBuilder](https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/x/auth/migrations/legacytx/stdtx_builder.go#L11-L17) for `SIGN_MODE_LEGACY_AMINO_JSON`.

However, the two implementations of `TxBuilder` should be hidden away from end-users, as they should prefer using the overarching `TxConfig` interface:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/client/tx_config.go#L27-L37
```

`TxConfig` is an app-wide configuration for managing transactions. Most importantly, it holds the information about whether to sign each transaction with `SIGN_MODE_DIRECT` or `SIGN_MODE_LEGACY_AMINO_JSON`. By calling `txBuilder := txConfig.NewTxBuilder()`, a new `TxBuilder` will be created with the appropriate sign mode.

Once `TxBuilder` is correctly populated with the setters exposed above, `TxConfig` will also take care of correctly encoding the bytes (again, either using `SIGN_MODE_DIRECT` or `SIGN_MODE_LEGACY_AMINO_JSON`). Here's a pseudo-code snippet of how to generate and encode a transaction, using the `TxEncoder()` method:

```go
txBuilder := txConfig.NewTxBuilder()
txBuilder.SetMsgs(...) // and other setters on txBuilder

bz, err := txConfig.TxEncoder()(txBuilder.GetTx())
// bz are bytes to be broadcasted over the network
```

### Broadcasting the Transaction

Once the transaction bytes are generated, there are currently three ways of broadcasting it.

#### CLI

Application developers create entry points to the application by creating a [command-line interface](./07-cli.md), [gRPC and/or REST interface](./06-grpc_rest.md), typically found in the application's `./cmd` folder. These interfaces allow users to interact with the application through command-line.

For the [command-line interface](../../build/building-modules/09-module-interfaces.md#cli), module developers create subcommands to add as children to the application top-level transaction command `TxCmd`. CLI commands actually bundle all the steps of transaction processing into one simple command: creating messages, generating transactions and broadcasting. For concrete examples, see the [Interacting with a Node](https://docs.cosmos.network/main/user/run-node/interact-node) section. An example transaction made using CLI looks like:

```bash
simd tx send $MY_VALIDATOR_ADDRESS $RECIPIENT 1000stake
```

#### gRPC

[gRPC](https://grpc.io) is the main component for the Cosmos SDK's RPC layer. Its principal usage is in the context of modules' [`Query` services](../../build/building-modules/04-query-services.md). However, the Cosmos SDK also exposes a few other module-agnostic gRPC services, one of them being the `Tx` service:

https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/proto/cosmos/tx/v1beta1/service.proto

The `Tx` service exposes a handful of utility functions, such as simulating a transaction or querying a transaction, and also one method to broadcast transactions.

Examples of broadcasting and simulating a transaction are shown [here](https://docs.cosmos.network/main/user/run-node/txs#programmatically-with-go).

#### REST

Each gRPC method has its corresponding REST endpoint, generated using [gRPC-gateway](https://github.com/grpc-ecosystem/grpc-gateway). Therefore, instead of using gRPC, you can also use HTTP to broadcast the same transaction, on the `POST /cosmos/tx/v1beta1/txs` endpoint.

An example can be seen [here](https://docs.cosmos.network/main/user/run-node/txs#using-rest)

#### CometBFT RPC

The three methods presented above are actually higher abstractions over the CometBFT RPC `/broadcast_tx_{async,sync,commit}` endpoints, documented [here](https://docs.cometbft.com/v1.0/explanation/core/rpc). This means that you can use the CometBFT RPC endpoints directly to broadcast the transaction, if you wish so.


#### Complete Example Building and Broadcasting a Transaction using clientv2

1. Correctly set up imports

```go
import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/client/v2/broadcast/comet"
	"cosmossdk.io/client/v2/tx"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/math"
	banktypes "cosmossdk.io/x/bank/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/codec"
	addrcodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

```

2. Create a gRPC connection

```go
clientConn, err := grpc.NewClient("127.0.0.1:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
	log.Fatal(err)
}
```

3. Setup codec and interface registry

```go
	// Setup interface registry and register necessary interfaces
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	banktypes.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)

	// Create a ProtoCodec for encoding/decoding
	protoCodec := codec.NewProtoCodec(interfaceRegistry)

```

4. Initialize keyring

```go

	ckr, err := keyring.New("autoclikeyring", "test", home, nil, protoCodec)
	if err != nil {
		log.Fatal("error creating keyring", err)
	}
	kr, err := keyring.NewAutoCLIKeyring(ckr, addrcodec.NewBech32Codec("cosmos"))
	if err != nil {
		log.Fatal("error creating auto cli keyring", err)
	}


```

5. Setup transaction parameters

```go

	// Setup transaction parameters
	txParams := tx.TxParameters{
		ChainID:  "simapp-v2-chain",
		SignMode: apisigning.SignMode_SIGN_MODE_DIRECT,
		AccountConfig: tx.AccountConfig{
			FromAddress: "cosmos1t0fmn0lyp2v99ga55mm37mpnqrlnc4xcs2hhhy",
			FromName:    "alice",
		},
	}

	// Configure gas settings
	gasConfig, err := tx.NewGasConfig(100, 100, "0stake")
	if err != nil {
		log.Fatal("error creating gas config: ", err)
	}
	txParams.GasConfig = gasConfig

	// Create auth query client
	authClient := authtypes.NewQueryClient(clientConn)

	// Retrieve account information for the sender
	fromAccount, err := getAccount("cosmos1t0fmn0lyp2v99ga55mm37mpnqrlnc4xcs2hhhy", authClient, protoCodec)
	if err != nil {
		log.Fatal("error getting from account: ", err)
	}

	// Update txParams with the correct account number and sequence
	txParams.AccountConfig.AccountNumber = fromAccount.GetAccountNumber()
	txParams.AccountConfig.Sequence = fromAccount.GetSequence()

	// Retrieve account information for the recipient
	toAccount, err := getAccount("cosmos1e2wanzh89mlwct7cs7eumxf7mrh5m3ykpsh66m", authClient, protoCodec)
	if err != nil {
		log.Fatal("error getting to account: ", err)
	}

	// Configure transaction settings
	txConf, _ := tx.NewTxConfig(tx.ConfigOptions{
		AddressCodec:          addrcodec.NewBech32Codec("cosmos"),
		Cdc:                   protoCodec,
		ValidatorAddressCodec: addrcodec.NewBech32Codec("cosmosval"),
		EnabledSignModes:      []apisigning.SignMode{apisigning.SignMode_SIGN_MODE_DIRECT},
	})
```

6. Build the transaction

```go
// Create a transaction factory
	f, err := tx.NewFactory(kr, codec.NewProtoCodec(codectypes.NewInterfaceRegistry()), nil, txConf, addrcodec.NewBech32Codec("cosmos"), clientConn, txParams)
	if err != nil {
		log.Fatal("error creating factory", err)
	}

	// Define the transaction message
	msgs := []transaction.Msg{
		&banktypes.MsgSend{
			FromAddress: fromAccount.GetAddress().String(),
			ToAddress:   toAccount.GetAddress().String(),
			Amount: sdk.Coins{
				sdk.NewCoin("stake", math.NewInt(1000000)),
			},
		},
	}

	// Build and sign the transaction
	tx, err := f.BuildsSignedTx(context.Background(), msgs...)
	if err != nil {
		log.Fatal("error building signed tx", err)
	}


```

7. Broadcast the transaction

```go
// Create a broadcaster for the transaction
	c, err := comet.NewCometBFTBroadcaster("http://127.0.0.1:26675", comet.BroadcastSync, protoCodec)
	if err != nil {
		log.Fatal("error creating comet broadcaster", err)
	}

	// Broadcast the transaction
	res, err := c.Broadcast(context.Background(), tx.Bytes())
	if err != nil {
		log.Fatal("error broadcasting tx", err)
	}

```

8. Helpers
    
```go
// getAccount retrieves account information using the provided address
func getAccount(address string, authClient authtypes.QueryClient, codec codec.Codec) (sdk.AccountI, error) {
	// Query account info
	accountQuery, err := authClient.Account(context.Background(), &authtypes.QueryAccountRequest{
		Address: string(address),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting account: %w", err)
	}

	// Unpack the account information
	var account sdk.AccountI
	err = codec.InterfaceRegistry().UnpackAny(accountQuery.Account, &account)
	if err != nil {
		return nil, fmt.Errorf("error unpacking account: %w", err)
	}

	return account, nil
}
```