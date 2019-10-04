# Encoding

## Pre-requisite Reading

- [Anatomy of an SDK application](../basics/app-anatomy.md)

## Synopsis

The `codec` is used everywhere in the Cosmos SDK to encode and decode structs and interfaces. The specific codec used in the Cosmos SDK is called `go-amino`. 

## Encoding

Every Cosmos SDK application exposes a global `codec` to marshal/unmarshal structs and interfaces in order to store and/or transfer them. As of now, the `codec` used in the Cosmos SDK is [go-amino](https://github.com/tendermint/go-amino), which possesses the following important properties:

- Interface support. 
- Deterministic encoding of value (which is required considering that blockchains are deterministic replicated state-machines). 
- Upgradeable schemas. 

The application's `codec` is typically initialized in the [application's constructor function](../basics/app-anatomy.md#constructor-function), where it is also passed to each of the application's modules via the [basic manager](../building-modules/module-manager.md#basic-manager). 

Among other things, the `codec` is used by module's [`keeper`s](../building-modules/keeper.md) to marshal objects into `[]byte` before storing them in the module's [`KVStore`](./store.md#kvstore), or to unmarshal them from `[]byte` when retrieving them:

```go
// typical pattern to marshal an object to []byte before storing it
bz := keeper.cdc.MustMarshalBinaryBare(object)

//typical pattern to unmarshal an object from []byte when retrieving it
keeper.cdc.MustUnmarshalBinaryBare(bz, &object)
```

Alternatively, it is possible to use `MustMarshalBinaryLengthPrefixed`/`MustUnmarshalBinaryLengthPrefixed` instead of `MustMarshalBinaryBare`/`MustUnmarshalBinaryBare` for the same encoding prefixed by a uvarint encoding of the object to encode. 

Another important use of the `codec` is the encoding and decoding of [transactions](./transactions.md). Transactions are defined at the Cosmos SDK level, but passed to the underlying consensus engine in order to be relayed to other peers. Since the underlying consensus engine is agnostic to the application, it only accepts transactions in the form of `[]byte`. The encoding is done by an object called [`TxEncoder`](https://github.com/cosmos/cosmos-sdk/blob/master/types/tx_msg.go#L48-L49) and the decoding by an object called [`TxDecoder`](https://github.com/cosmos/cosmos-sdk/blob/master/types/tx_msg.go#L45-L47). A [standard implementation](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/types/stdtx.go) of both these objects can be found in the `auth` module:

```go
// DefaultTxDecoder logic for standard transaction decoding
func DefaultTxDecoder(cdc *codec.Codec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx = StdTx{}

		if len(txBytes) == 0 {
			return nil, sdk.ErrTxDecode("txBytes are empty")
		}

		// StdTx.Msg is an interface. The concrete types
		// are registered by MakeTxCodec
		err := cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
		if err != nil {
			return nil, sdk.ErrTxDecode("error decoding transaction").TraceSDK(err.Error())
		}

		return tx, nil
	}
}

// DefaultTxEncoder logic for standard transaction encoding
func DefaultTxEncoder(cdc *codec.Codec) sdk.TxEncoder {
	return func(tx sdk.Tx) ([]byte, error) {
		return cdc.MarshalBinaryLengthPrefixed(tx)
	}
}
```

## Next

Learn about next concept 