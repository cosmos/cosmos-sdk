---
sidebar_position: 1
---

# `x/auth`

## Abstract

This document specifies the auth module of the Cosmos SDK.

The auth module is responsible for specifying the base transaction and account types
for an application, since the SDK itself is agnostic to these particulars. It contains
the middlewares, where all basic transaction validity checks (signatures, nonces, auxiliary fields)
are performed, and exposes the account keeper, which allows other modules to read, write, and modify accounts.

This module is used in the Cosmos Hub.

## Contents

* [Concepts](#concepts)
    * [Gas & Fees](#gas--fees)
* [State](#state)
    * [Accounts](#accounts)
    * [Account Rekeying State](#account-rekeying-state)
* [AnteHandlers](#antehandlers)
* [Messages](#messages)
    * [MsgChangePubKey](#msgchangepubkey)
* [Keepers](#keepers)
    * [Account Keeper](#account-keeper)
* [Parameters](#parameters)
* [Events](#events)
* [Client](#client)
    * [CLI](#cli)
    * [gRPC](#grpc)
    * [REST](#rest)

## Concepts

**Note:** The auth module is different from the [authz module](../authz/).

The differences are:

* `auth` - authentication of accounts and transactions for Cosmos SDK applications and is responsible for specifying the base transaction and account types.
* `authz` - authorization for accounts to perform actions on behalf of other accounts and enables a granter to grant authorizations to a grantee that allows the grantee to execute messages on behalf of the granter.

### Gas & Fees

Fees serve two purposes for an operator of the network.

Fees limit the growth of the state stored by every full node and allow for
general purpose censorship of transactions of little economic value. Fees
are best suited as an anti-spam mechanism where validators are disinterested in
the use of the network and identities of users.

Fees are determined by the gas limits and gas prices transactions provide, where
`fees = ceil(gasLimit * gasPrices)`. Txs incur gas costs for all state reads/writes,
signature verification, as well as costs proportional to the tx size. Operators
should set minimum gas prices when starting their nodes. They must set the unit
costs of gas in each token denomination they wish to support:

`simd start ... --minimum-gas-prices=0.00001stake;0.05photinos`

When adding transactions to mempool or gossiping transactions, validators check
if the transaction's gas prices, which are determined by the provided fees, meet
any of the validator's minimum gas prices. In other words, a transaction must
provide a fee of at least one denomination that matches a validator's minimum
gas price.

CometBFT does not currently provide fee based mempool prioritization, and fee
based mempool filtering is local to node and not part of consensus. But with
minimum gas prices set, such a mechanism could be implemented by node operators.

Because the market value for tokens will fluctuate, validators are expected to
dynamically adjust their minimum gas prices to a level that would encourage the
use of the network.		

## State

### Accounts

Accounts contain authentication information for a uniquely identified external user of an SDK blockchain,
including public key, address, and account number / sequence number for replay protection. For efficiency,
since account balances must also be fetched to pay fees, account structs also store the balance of a user
as `sdk.Coins`.

Accounts are exposed externally as an interface, and stored internally as
either a base account or vesting account. Module clients wishing to add more
account types may do so.

* `0x01 | Address -> ProtocolBuffer(account)`

#### Account Interface

The account interface exposes methods to read and write standard account information.
Note that all of these methods operate on an account struct conforming to the
interface - in order to write the account to the store, the account keeper will
need to be used.

```go
// AccountI is an interface used to store coins at a given address within state.
// It presumes a notion of sequence numbers for replay protection,
// a notion of account numbers for replay protection for previously pruned accounts,
// and a pubkey for authentication purposes.
//
// Many complex conditions can be used in the concrete struct which implements AccountI.
type AccountI interface {
	proto.Message

	GetAddress() sdk.AccAddress
	SetAddress(sdk.AccAddress) error // errors if already set.

	GetPubKey() crypto.PubKey // can return nil.
	SetPubKey(crypto.PubKey) error

	GetAccountNumber() uint64
	SetAccountNumber(uint64) error

	GetSequence() uint64
	SetSequence(uint64) error

	// Ensure that account implements stringer
	String() string
}
```

##### Base Account

A base account is the simplest and most common account type, which just stores all requisite
fields directly in a struct.

```protobuf
// BaseAccount defines a base account type. It contains all the necessary fields
// for basic account functionality. Any custom account type should extend this
// type for additional functionality (e.g. vesting).
message BaseAccount {
  string address = 1;
  google.protobuf.Any pub_key = 2;
  uint64 account_number = 3;
  uint64 sequence       = 4;
}
```

### Vesting Account

See [Vesting](https://docs.cosmos.network/main/modules/auth/vesting/).

### Account Rekeying State

An account can replace its public key with [`MsgChangePubKey`](#msgchangepubkey) and keep its address. After that,
an account's address is no longer always the hash of its public key. Two more collections in the account store
record rotations:

* PubKeyHistory: `0x5b | Address | BigEndian(rotationIndex) -> ProtocolBuffer(PubKeyHistoryEntry)`. There is one entry
  per rotation, keyed by a per-account rotation index (a `uint64` counting from 0). It holds the replaced pubkey, the
  block height (`replaced_at_height`) and block time it was replaced, and the natural address of the key that
  replaced it.
* RekeyIndex: `0x5c | NaturalAddress | Address -> []byte{}`. `NaturalAddress` is the hash of the account's current
  pubkey. An entry exists only while that hash differs from the account address, so rotating back to the original
  key removes it.

Genesis exports the history as `pub_key_history`. `InitGenesis` rebuilds the index from the accounts' stored
pubkeys. `ValidateGenesis` accepts an account whose pubkey does not hash to its address only if `pub_key_history`
has an entry for that address. Module credential accounts are exempt. Every `pub_key_history` entry must belong to an
account in genesis whose pubkey is neither nil nor a module credential, and each history must form a chain: an entry's
`new_key_address` is the natural address of the next entry's pubkey, and the last entry's is the natural address of the
account's current pubkey. No entry's pubkey may be a module credential. `RemoveAccount` deletes the account's history
and index entry along with the account, so an export never contains history without an account.

## AnteHandlers

Besides its [messages](#messages), the `x/auth` module exposes the special `AnteHandler`, used for performing basic validity checks on a transaction, such that it could be thrown out of the mempool.
The `AnteHandler` can be seen as a set of decorators that check transactions within the current context, per [ADR 010](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-010-modular-antehandler.md).

Note that the `AnteHandler` is called on both `CheckTx` and `DeliverTx`, as CometBFT proposers presently have the ability to include in their proposed block transactions which fail `CheckTx`.

### Decorators

The auth module provides `AnteDecorator`s that are recursively chained together into a single `AnteHandler` in the following order:

* `SetUpContextDecorator`: Sets the `GasMeter` in the `Context` and wraps the next `AnteHandler` with a defer clause to recover from any downstream `OutOfGas` panics in the `AnteHandler` chain to return an error with information on gas provided and gas used.

* `RejectExtensionOptionsDecorator`: Rejects all extension options which can optionally be included in protobuf transactions.

* `MempoolFeeDecorator`: Checks if the `tx` fee is above local mempool `minFee` parameter during `CheckTx`.

* `ValidateBasicDecorator`: Calls `tx.ValidateBasic` and returns any non-nil error.

* `TxTimeoutHeightDecorator`: Check for a `tx` height timeout.

* `ValidateMemoDecorator`: Validates `tx` memo with application parameters and returns any non-nil error.

* `ConsumeGasTxSizeDecorator`: Consumes gas proportional to the `tx` size based on application parameters.

* `DeductFeeDecorator`: Deducts the `FeeAmount` from first signer of the `tx`. If the `x/feegrant` module is enabled and a fee granter is set, it deducts fees from the fee granter account.

* `SetPubKeyDecorator`: Sets the pubkey from a `tx`'s signers that does not already have its corresponding pubkey saved in the state machine and in the current context. A signer without a stored pubkey must provide a pubkey whose address equals the signer address. If a signer already has a stored pubkey, which may have been rotated with `MsgChangePubKey`, any pubkey in the `tx` must equal the stored one.

* `ValidateSigCountDecorator`: Validates the number of signatures in `tx` based on app-parameters.

* `SigGasConsumeDecorator`: Consumes parameter-defined amount of gas for each signature. This requires pubkeys to be set in context for all signers as part of `SetPubKeyDecorator`.

* `SigVerificationDecorator`: Verifies all signatures are valid. This requires pubkeys to be set in context for all signers as part of `SetPubKeyDecorator`.

* `IncrementSequenceDecorator`: Increments the account sequence for each signer to prevent replay attacks.

## Messages

### MsgUpdateParams

Updates the module parameters. The authority, by default the governance module account, must sign it. It replaces
all parameters at once.

### MsgChangePubKey

Replaces the public key of an account. The account keeps its address, account number, sequence, balances and
any validator it operates. Signed by the account's current key, it changes the members or threshold of a
multisig account, or moves an account to another key type, such as ML-DSA-65, without moving funds.

```protobuf
message MsgChangePubKey {
  option (cosmos.msg.v1.signer) = "address";

  string              address     = 1;
  google.protobuf.Any new_pub_key = 2;
  bytes               proof       = 3;
}
```

`proof` is a proof of possession of `new_pub_key`. It is the proto encoding of a
`cosmos.tx.signing.v1beta1.SignatureDescriptor.Data` that signs an
[ADR-036](../../docs/architecture/adr-036-arbitrary-signature.md) amino-JSON sign doc. In that doc, `signer` is the
bech32 address of `new_pub_key` and `data` is the proto encoding of
`ChangePubKeyProofDoc{chain_id, account_number, address, new_pub_key}`. Every signature in the proof must use
`SIGN_MODE_LEGACY_AMINO_JSON`, which Ledger devices support. For a multisig `new_pub_key`, the proof is a
multisignature that meets its threshold. The doc binds the chain id, account number and address, so a proof cannot be
replayed for another account or chain.

The handler checks, in order:

1. `pub_key_change_enabled` is `true` (`ErrPubKeyChangeDisabled`).
2. The account exists, is not a module account, has a stored pubkey, and that pubkey is not a `ModuleCredential`
   (`ErrAccountNotRekeyable`).
3. `new_pub_key` is supported (`ErrInvalidNewPubKey`). Supported types are `secp256k1`, `secp256r1`, `ed25519`,
   `mldsa65`, and `LegacyAminoPubKey` multisigs built from `secp256k1`, `ed25519` and `mldsa65` keys. A multisig may
   contain multisigs, but those may not (at most 2 levels of multisig), and each multisig may have at most 32 direct
   subkeys: these are the limits the ante handler accepts when flattening signatures. `secp256r1` is accepted only as
   a top-level key. The total number of subkeys must not exceed `TxSigLimit`, so the account cannot lock itself out,
   and the key must differ from the current one.
4. It consumes `pub_key_change_cost` gas, then verifies the proof (`ErrInvalidPubKeyProof`).

On success, it records the replaced key in the pubkey history, updates the rekey index, stores the new pubkey and
emits a [`change_pubkey`](#events) event. From then on, the account's txs must be signed with the new key. Txs signed
with the old key fail in `FinalizeBlock` and do not change state. One that includes the old pubkey also fails
`SetPubKeyDecorator` on `RecheckTx` and is evicted from the mempool. One that omits its pubkey passes, and `RecheckTx`
skips signature verification, so it can stay in the mempool until it expires or is included and fails.

`x/authz` will not grant or execute `MsgChangePubKey`, because that would let the grantee take over the granter's
account.

## Keepers

The auth module only exposes one keeper, the account keeper, which can be used to read and write accounts.

### Account Keeper

Presently only one fully-permissioned account keeper is exposed, which has the ability to both read and write
all fields of all accounts, and to iterate over all stored accounts.

```go
// AccountKeeperI is the interface contract that x/auth's keeper implements.
type AccountKeeperI interface {
	// Return a new account with the next account number and the specified address. Does not save the new account to the store.
	NewAccountWithAddress(context.Context, sdk.AccAddress) sdk.AccountI

	// Return a new account with the next account number. Does not save the new account to the store.
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI

	// Check if an account exists in the store.
	HasAccount(context.Context, sdk.AccAddress) bool

	// Retrieve an account from the store.
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI

	// Set an account in the store.
	SetAccount(context.Context, sdk.AccountI)

	// Remove an account from the store.
	RemoveAccount(context.Context, sdk.AccountI)

	// Iterate over all accounts, calling the provided function. Stop iteration when it returns true.
	IterateAccounts(context.Context, func(sdk.AccountI) bool)

	// Fetch the public key of an account at a specified address
	GetPubKey(context.Context, sdk.AccAddress) (cryptotypes.PubKey, error)

	// Fetch the sequence of an account at a specified address.
	GetSequence(context.Context, sdk.AccAddress) (uint64, error)

	// Fetch the next account number, and increment the internal counter.
	NextAccountNumber(context.Context) uint64

	// GetModulePermissions fetches per-module account permissions
	GetModulePermissions() map[string]types.PermissionsForAddress

	// AddressCodec returns the account address codec.
	AddressCodec() address.Codec
}
```

## Parameters

The auth module contains the following parameters:

| Key                    | Type            | Example |
| ---------------------- | --------------- | ------- |
| MaxMemoCharacters      |      uint64     | 256     |
| TxSigLimit             |      uint64     | 7       |
| TxSizeCostPerByte      |      uint64     | 10      |
| SigVerifyCostED25519   |      uint64     | 590     |
| SigVerifyCostSecp256k1 |      uint64     | 1000    |
| PubKeyChangeEnabled    |      bool       | false   |
| PubKeyChangeCost       |      uint64     | 50000   |

`PubKeyChangeEnabled` turns `MsgChangePubKey` on. It defaults to `false`, and the 7 to 8 store migration sets it to
`false`. `PubKeyChangeCost` is the gas that `MsgChangePubKey` consumes in addition to the normal tx gas.

## Events

### MsgChangePubKey

| Type          | Attribute Key      | Attribute Value                        |
| ------------- | ------------------ | -------------------------------------- |
| change_pubkey | address            | {accountAddress}                       |
| change_pubkey | old_pubkey_address | {naturalAddressOfReplacedPubKey}       |
| change_pubkey | new_pubkey_address | {naturalAddressOfNewPubKey}            |

## Client

### CLI

A user can query and interact with the `auth` module using the CLI.

### Query

The `query` commands allow users to query `auth` state.

```bash
simd query auth --help
```

#### account

The `account` command allows users to query for an account by its address.

```bash
simd query auth account [address] [flags]
```

Example:

```bash
simd query auth account cosmos1...
```

Example Output:

```bash
'@type': /cosmos.auth.v1beta1.BaseAccount
account_number: "0"
address: cosmos1zwg6tpl8aw4rawv8sgag9086lpw5hv33u5ctr2
pub_key:
  '@type': /cosmos.crypto.secp256k1.PubKey
  key: ApDrE38zZdd7wLmFS9YmqO684y5DG6fjZ4rVeihF/AQD
sequence: "1"
```

#### accounts

The `accounts` command allow users to query all the available accounts.

```bash
simd query auth accounts [flags]
```

Example:

```bash
simd query auth accounts
```

Example Output:

```bash
accounts:
- '@type': /cosmos.auth.v1beta1.BaseAccount
  account_number: "0"
  address: cosmos1zwg6tpl8aw4rawv8sgag9086lpw5hv33u5ctr2
  pub_key:
    '@type': /cosmos.crypto.secp256k1.PubKey
    key: ApDrE38zZdd7wLmFS9YmqO684y5DG6fjZ4rVeihF/AQD
  sequence: "1"
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "8"
    address: cosmos1yl6hdjhmkf37639730gffanpzndzdpmhwlkfhr
    pub_key: null
    sequence: "0"
  name: transfer
  permissions:
  - minter
  - burner
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "4"
    address: cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh
    pub_key: null
    sequence: "0"
  name: bonded_tokens_pool
  permissions:
  - burner
  - staking
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "5"
    address: cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r
    pub_key: null
    sequence: "0"
  name: not_bonded_tokens_pool
  permissions:
  - burner
  - staking
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "6"
    address: cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn
    pub_key: null
    sequence: "0"
  name: gov
  permissions:
  - burner
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "3"
    address: cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl
    pub_key: null
    sequence: "0"
  name: distribution
  permissions: []
- '@type': /cosmos.auth.v1beta1.BaseAccount
  account_number: "1"
  address: cosmos147k3r7v2tvwqhcmaxcfql7j8rmkrlsemxshd3j
  pub_key: null
  sequence: "0"
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "7"
    address: cosmos1m3h30wlvsf8llruxtpukdvsy0km2kum8g38c8q
    pub_key: null
    sequence: "0"
  name: mint
  permissions:
  - minter
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "2"
    address: cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta
    pub_key: null
    sequence: "0"
  name: fee_collector
  permissions: []
pagination:
  next_key: null
  total: "0"
```

#### params

The `params` command allow users to query the current auth parameters.

```bash
simd query auth params [flags]
```

Example:

```bash
simd query auth params
```

Example Output:

```bash
max_memo_characters: "256"
sig_verify_cost_ed25519: "590"
sig_verify_cost_secp256k1: "1000"
tx_sig_limit: "7"
tx_size_cost_per_byte: "10"
```

#### rekeyed-accounts

The `rekeyed-accounts` command lists the accounts whose current pubkey hashes to the given address. After a
`MsgChangePubKey`, a key controls an account whose address differs from the key's own address. Wallets that recover a
key from a mnemonic use this command to find that account.

```bash
simd query auth rekeyed-accounts [address] [flags]
```

#### pubkey-history

The `pubkey-history` command lists the pubkeys an account has replaced, with the height and time of each rotation.

```bash
simd query auth pubkey-history [address] [flags]
```

### Transactions

The `auth` module supports transactions commands to help you with signing and more. Compared to other modules you can access directly the `auth` module transactions commands using the only `tx` command.

Use directly the `--help` flag to get more information about the `tx` command.

```bash
simd tx --help
```

#### `sign`

The `sign` command allows users to sign transactions that was generated offline.

```bash
simd tx sign tx.json --from $ALICE > tx.signed.json
```

The result is a signed transaction that can be broadcasted to the network thanks to the broadcast command.

More information about the `sign` command can be found running `simd tx sign --help`.

#### `sign-batch`

The `sign-batch` command allows users to sign multiples offline generated transactions.
The transactions can be in one file, with one tx per line, or in multiple files.

```bash
simd tx sign txs.json --from $ALICE > tx.signed.json
```

or

```bash 
simd tx sign tx1.json tx2.json tx3.json --from $ALICE > tx.signed.json
```

The result is multiples signed transactions. For combining the signed transactions into one transactions, use the `--append` flag.

More information about the `sign-batch` command can be found running `simd tx sign-batch --help`.

#### `multi-sign`

The `multi-sign` command allows users to sign transactions that was generated offline by a multisig account.

```bash
simd tx multisign transaction.json k1k2k3 k1sig.json k2sig.json k3sig.json
```

Where `k1k2k3` is the multisig account address, `k1sig.json` is the signature of the first signer, `k2sig.json` is the signature of the second signer, and `k3sig.json` is the signature of the third signer.

##### Nested multisig transactions

To allow transactions to be signed by nested multisigs, meaning that a participant of a multisig account can be another multisig account, the `--skip-signature-verification` flag must be used.

```bash
# First aggregate signatures of the multisig participant
simd tx multi-sign transaction.json ms1 ms1p1sig.json ms1p2sig.json --signature-only --skip-signature-verification > ms1sig.json

# Then use the aggregated signatures and the other signatures to sign the final transaction
simd tx multi-sign transaction.json k1ms1 k1sig.json ms1sig.json --skip-signature-verification
```

Where `ms1` is the nested multisig account address, `ms1p1sig.json` is the signature of the first participant of the nested multisig account, `ms1p2sig.json` is the signature of the second participant of the nested multisig account, and `ms1sig.json` is the aggregated signature of the nested multisig account.

`k1ms1` is a multisig account comprised of an individual signer and another nested multisig account (`ms1`). `k1sig.json` is the signature of the first signer of the individual member.

More information about the `multi-sign` command can be found running `simd tx multi-sign --help`.

#### `multisign-batch`

The `multisign-batch` works the same way as `sign-batch`, but for multisig accounts.
With the difference that the `multisign-batch` command requires all transactions to be in one file, and the `--append` flag does not exist.

More information about the `multisign-batch` command can be found running `simd tx multisign-batch --help`.

#### `validate-signatures`

The `validate-signatures` command allows users to validate the signatures of a signed transaction.

```bash
$ simd tx validate-signatures tx.signed.json
Signers:
  0: cosmos1l6vsqhh7rnwsyr2kyz3jjg3qduaz8gwgyl8275

Signatures:
  0: cosmos1l6vsqhh7rnwsyr2kyz3jjg3qduaz8gwgyl8275                      [OK]
```

More information about the `validate-signatures` command can be found running `simd tx validate-signatures --help`.

#### `broadcast`

The `broadcast` command allows users to broadcast a signed transaction to the network.

```bash
simd tx broadcast tx.signed.json
```

More information about the `broadcast` command can be found running `simd tx broadcast --help`.


#### `sign-rekey-proof`

The `sign-rekey-proof` command signs the proof of possession that `change-pubkey` needs from the new pubkey. The
`--from` key must be the new pubkey or, for a multisig new pubkey, one of its direct members. Each member signs its
own proof. The proof is always signed with `SIGN_MODE_LEGACY_AMINO_JSON`.

```bash
simd tx auth sign-rekey-proof [account-address] [new-pubkey-json] --from <new-key> > proof.json
```

#### `change-pubkey`

The `change-pubkey` command sends a `MsgChangePubKey` signed by the account's current key. `--proof` takes one or
more files written by `sign-rekey-proof`, comma separated. For a multisig new pubkey, they are combined into one
multisignature.

```bash
simd tx auth change-pubkey [account-address] [new-pubkey-json] --proof proof1.json,proof2.json --from <current-key>
```

After the change, sign the account's txs with the new key and pass the account address with `--signer-address`,
because the address of the `--from` key no longer matches the account. For commands such as `bank send` that take
the sender as an argument, pass the new key's name there:

```bash
simd tx bank send <new-key> [to-address] 10stake --signer-address [account-address]
```

An account rekeyed to a multisig spends with the usual multisig flow, passing `--signer-address` to each step, so the
signatures are made for the account rather than for the multisig key's own address. As for any multisig, only
`SIGN_MODE_LEGACY_AMINO_JSON` is supported:

```bash
simd tx sign tx.json --from <member-key> --multisig <multisig-key> --signer-address [account-address] --sign-mode amino-json > sig.json
simd tx multisign tx.json <multisig-key> sig1.json sig2.json --signer-address [account-address] > signed.json
```

`sign-batch --multisig` and `multisign-batch` accept `--signer-address` in the same way.

### gRPC

A user can query the `auth` module using gRPC endpoints.

#### Account

The `account` endpoint allows users to query for an account by its address.

```bash
cosmos.auth.v1beta1.Query/Account
```

Example:

```bash
grpcurl -plaintext \
    -d '{"address":"cosmos1.."}' \
    localhost:9090 \
    cosmos.auth.v1beta1.Query/Account
```

Example Output:

```bash
{
  "account":{
    "@type":"/cosmos.auth.v1beta1.BaseAccount",
    "address":"cosmos1zwg6tpl8aw4rawv8sgag9086lpw5hv33u5ctr2",
    "pubKey":{
      "@type":"/cosmos.crypto.secp256k1.PubKey",
      "key":"ApDrE38zZdd7wLmFS9YmqO684y5DG6fjZ4rVeihF/AQD"
    },
    "sequence":"1"
  }
}
```

#### Accounts

The `accounts` endpoint allow users to query all the available accounts.

```bash
cosmos.auth.v1beta1.Query/Accounts
```

Example:

```bash
grpcurl -plaintext \
    localhost:9090 \
    cosmos.auth.v1beta1.Query/Accounts
```

Example Output:

```bash
{
   "accounts":[
      {
         "@type":"/cosmos.auth.v1beta1.BaseAccount",
         "address":"cosmos1zwg6tpl8aw4rawv8sgag9086lpw5hv33u5ctr2",
         "pubKey":{
            "@type":"/cosmos.crypto.secp256k1.PubKey",
            "key":"ApDrE38zZdd7wLmFS9YmqO684y5DG6fjZ4rVeihF/AQD"
         },
         "sequence":"1"
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos1yl6hdjhmkf37639730gffanpzndzdpmhwlkfhr",
            "accountNumber":"8"
         },
         "name":"transfer",
         "permissions":[
            "minter",
            "burner"
         ]
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
            "accountNumber":"4"
         },
         "name":"bonded_tokens_pool",
         "permissions":[
            "burner",
            "staking"
         ]
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r",
            "accountNumber":"5"
         },
         "name":"not_bonded_tokens_pool",
         "permissions":[
            "burner",
            "staking"
         ]
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
            "accountNumber":"6"
         },
         "name":"gov",
         "permissions":[
            "burner"
         ]
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl",
            "accountNumber":"3"
         },
         "name":"distribution"
      },
      {
         "@type":"/cosmos.auth.v1beta1.BaseAccount",
         "accountNumber":"1",
         "address":"cosmos147k3r7v2tvwqhcmaxcfql7j8rmkrlsemxshd3j"
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos1m3h30wlvsf8llruxtpukdvsy0km2kum8g38c8q",
            "accountNumber":"7"
         },
         "name":"mint",
         "permissions":[
            "minter"
         ]
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta",
            "accountNumber":"2"
         },
         "name":"fee_collector"
      }
   ],
   "pagination":{
      "total":"9"
   }
}
```

#### Params

The `params` endpoint allow users to query the current auth parameters.

```bash
cosmos.auth.v1beta1.Query/Params
```

Example:

```bash
grpcurl -plaintext \
    localhost:9090 \
    cosmos.auth.v1beta1.Query/Params
```

Example Output:

```bash
{
  "params": {
    "maxMemoCharacters": "256",
    "txSigLimit": "7",
    "txSizeCostPerByte": "10",
    "sigVerifyCostEd25519": "590",
    "sigVerifyCostSecp256k1": "1000"
  }
}
```

#### RekeyedAccounts

The `RekeyedAccounts` endpoint returns the accounts whose current pubkey hashes to the given address.

```bash
cosmos.auth.v1beta1.Query/RekeyedAccounts
```

Example:

```bash
grpcurl -plaintext \
    -d '{"address":"cosmos1.."}' \
    localhost:9090 \
    cosmos.auth.v1beta1.Query/RekeyedAccounts
```

#### PubKeyHistory

The `PubKeyHistory` endpoint returns the pubkeys an account has replaced.

```bash
cosmos.auth.v1beta1.Query/PubKeyHistory
```

Example:

```bash
grpcurl -plaintext \
    -d '{"address":"cosmos1.."}' \
    localhost:9090 \
    cosmos.auth.v1beta1.Query/PubKeyHistory
```

### REST

A user can query the `auth` module using REST endpoints.

#### Account

The `account` endpoint allows users to query for an account by its address.

```bash
/cosmos/auth/v1beta1/account?address={address}
```

#### Accounts

The `accounts` endpoint allows users to query all the available accounts.

```bash
/cosmos/auth/v1beta1/accounts
```

#### Params

The `params` endpoint allows users to query the current auth parameters.

```bash
/cosmos/auth/v1beta1/params
```

#### RekeyedAccounts

The `rekeyed_accounts` endpoint returns the accounts whose current pubkey hashes to the given address.

```bash
/cosmos/auth/v1beta1/rekeyed_accounts/{address}
```

#### PubKeyHistory

The `pub_key_history` endpoint returns the pubkeys an account has replaced.

```bash
/cosmos/auth/v1beta1/pub_key_history/{address}
```
