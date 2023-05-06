# RFC 004: Accounts

## Changelog

- 17/03/2023: DRAFT

## Abstract

This RFC presents a proposal which introduces a new account abstraction system to the Cosmos SDK. The proposal's main goals are:

Unifying ModuleAccounts and Externally Owned Accounts (EOA) to create a single actor type from a runtime perspective.
Developing a versatile account abstraction system that enables accounts to possess arbitrary logic concerning:
1. State transition verification (authentication)
2. Internal state modification
3. State transition execution (when modules communicate through messages rather than keepers)

Supporting additional use cases, such as:
1. More intelligent accounts
2. Streamlined extensions of authentication mechanisms (custom cryptographic pairings, but not only)
3. Vesting accounts

The proposal's full potential is realized when integrated with modules like `x/wasm` and `x/evm`. These modules facilitate 
the deployment of accounts with custom and dynamic code, eliminating the need for app-level definition of account code. This enables the
proposed account abstraction's use cases to evolve with the network itself, without needing to undergo chain upgrades.

## Proposal


### Account role

The proposal begins by clarifying the role of an account within the Cosmos SDK's runtime. Like modules, accounts are
actors that can send and receive messages. They modify each other's states through state transitions, which are represented
as messages. The runtime does not differentiate between actors or apply any authorization logic; it simply ensures that
sender actors cannot impersonate others.

This leads to significant implications: Externally Owned Accounts (EOA) and ModuleAccounts (system-owned identities) can
be treated as the same entity (from the runtime point of view). Authorization, which determines which actor can execute
a particular state transition towards another actor, must be managed by the recipient of the state transition. 
For example, the 'bank' module decides if the 'mint' module is permitted to inflate a coin's supply. This eliminates the
need for permissions in the 'auth' module as they currently exist.

The key difference between EOA and ModuleAccount lies in the origin of state transitions on their behalf. In the case of
ModuleAccount, a state transition cannot be initiated from a transaction (TX), as no valid credential can prove the sender
is the ModuleAccount itself.

We can distinguish between EOA and ModuleAccount by considering an account as a ModuleAccount (or SystemOwnedAccount)
when it is not possible to initiate a state transition on its behalf from a TX. Given that this distinction occurs during
the VerifySignature stage of the AnteHandler and that authorization is managed by the state transition recipient, the
execution runtime or other system actors to manage differences between account types. Instead, they should be treated as
business logic units capable of sending and receiving messages. This further proves that there's no need to distinguish
`ModuleAccounts` from `EOA`.

### Account Definition

We define the new `Account` type, which is what an account needs to implement to be treated as such.
An `Account` type is defined at APP level, so it cannot be dynamically loaded as the chain is running without upgrading the 
node code, unless we use something like `x/wasm`. 

```go
type Schema struct {
	state StateSchema // represents the state of an account
	init InitSchema // represents the init msg schema
	exec ExecSchema // represents the multiple execution msg schemas, containing also responses
	query QuerySchema // represents the multiple query msg schemas, containing also responses
	migrate *MigrateSchema // represents the multiple migrate msg schemas, containing also responses, it's optional
}

type InternalAccount struct {
	init    func(ctx *Context, msg proto.Message) (*InitResponse, error)
	execute func(ctx *Context, msg proto.Message) (*ExecuteResponse, error)
	query   func(ctx *Context, msg proto.Message) (proto.Message, error)
    schema  func() *Schema
    migrate func(ctx *Context, msg proto.Message) (*MigrateResponse, error)
}
```

This is an internal view of the account as intended by the system. It is not meant to be what developers implement. An
example implementation of the `InternalAccount` type can be found in [this](https://github.com/testinginprod/accounts-poc/blob/main/examples/recover/recover.go)
example of account whose credentials can be recovered. In fact, even if the `Internal` implementation is untyped (with
respect to `proto.Message`), the concrete implementation is fully typed.

During any of the execution methods of `InternalAccount`, `schema` excluded, the account is given a `Context` which provides:
- A namespaced `KVStore` for the account, which isolates the account state from others (NOTE: no `store keys` needed,
the account address serves as `store key`).
- Information regarding itself (its address)
- Information regarding the sender.
- ...

#### Init

Init defines the entrypoint that allows for a new account instance of a given kind to be initialised. 
The account is passed some opaque protobuf message which is then interpreted and contains the instructions that
constitute the initial state of an account once it is deployed.

An `Account` code can be deployed multiple times through the `Init` function, similar to how a `CosmWasm` contract code
can be deployed (Instantiated) multiple times.

#### Execute

Execute defines the entrypoint that allows an `Account` to process a state transition, the account can decide then how to
process the state transition based on the message provided and the sender of the transition.

#### Query

Query defines a read-only entrypoint that provides a stable interface that links an account with its state. The reason for 
which `Query` is still being preferred as an addition to raw state reflection is to:
- Provide a stable interface for querying (state can be optimised and change more frequently than a query)
- Provide a way to define an account `Interface` with respect to its `Read/Write` paths.
- Provide a way to query information that cannot be processed from raw state reflection, ex: compute information from lazy
state that has not been yet concretely processed (eg: balances with respect to lazy inputs/outputs)

#### Schema

Schema provides the definition of an account from `API` perspective, and it's the only thing that should be taken into account
when interacting with an account from another account or module, for example: an account is an `authz-interface` account if
it has the following message in its execution messages `MsgProxyStateTransition{ state_transition: google.Protobuf.Any }`.

### Migrate

Migrate defines the entrypoint that allows an `Account` to migrate its state from a previous version to a new one. Migrations
can be initiated only by the account itself, concretely this means that the migrate action sender can only be the account address
itself, if the account wants to allow another address to migrate it on its behalf then it could create an execution message
that makes the account migrate itself.

### x/accounts module

In order to create accounts we define a new module `x/accounts`, note that `x/accounts` deploys account with no authentication
credentials attached to it which means no action of an account can be incepted from a TX, we will later explore how the
`x/authn` module uses `x/accounts` to deploy authenticated accounts.

This also has another important implication for which account addresses are now fully decoupled from the authentication mechanism
which makes in turn off-chain operations a little more complex, as the chain becomes the real link between account identifier
and credentials. 

We could also introduce a way to deterministically compute the account address.

Note, from the transaction point of view, the `init_message` and `execute_message` are opaque `google.Protobuf.Any`.

The module protobuf definition for `x/accounts` are the following:

```protobuf
// Msg defines the Msg service.
service Msg {
  rpc Deploy(MsgDeploy) returns (MsgDeployResponse);
  rpc Execute(MsgExecute) returns (MsgExecuteResponse);
  rpc Migrate(MsgMigrate) returns (MsgMigrateResponse);
}

message MsgDeploy {
  string sender = 1;
  string kind = 2;
  google.Protobuf.Any init_message = 3;
  repeated google.Protobuf.Any authorize_messages = 4 [(gogoproto.nullable) = false];
}

message MsgDeployResponse {
  string address = 1;
  uint64 id = 2;
  google.Protobuf.Any data = 3;
}

message MsgExecute {
  string sender = 1;
  string address = 2;
  google.Protobuf.Any message = 3;
  repeated google.Protobuf.Any authorize_messages = 4 [(gogoproto.nullable) = false];
}

message MsgExecuteResponse {
  google.Protobuf.Any data = 1;
}

message MsgMigrate {
  string sender = 1;
  string new_account_kind = 2;
  google.Protobuf.Any migrate_message = 3;
}

message MsgMigrateResponse {
  google.Protobuf.Any data = 1;
}

```

#### MsgDeploy

Deploys a new instance of the given account `kind` with initial settings represented by the `init_message` which is a `google.Protobuf.Any`.
Of course the `init_message` can be empty. A response is returned containing the account ID and humanised address, alongside some response
that the account instantiation might produce.

#### Address derivation

In order to decouple public keys from account addresses, we introduce a new address derivation mechanism which is 


#### MsgExecute

Sends a `StateTransition` execution request, where the state transition is represented by the `message` which is a `google.Protobuf.Any`.
The account can then decide if to process it or not based on the `sender`.

### MsgMigrate

Migrates an account to a new version of itself, the new version is represented by the `new_account_kind`. The state transition
can only be incepted by the account itself, which means that the `sender` must be the account address itself. During the migration
the account current state is given to the new version of the account, which then executes the migration logic using the `migrate_message`,
it might change state or not, it's up to the account to decide. The response contains possible data that the account might produce
after the migration.

#### Authorize Messages

The `Deploy` and `Execute` messages have a field in common called `authorize_messages`, these messages are messages that the account
can execute on behalf of the sender. For example, in case an account is expecting some funds to be sent from the sender,
the sender can attach a `MsgSend` that the account can execute on the sender's behalf. These authorizations are short-lived,
they live only for the duration of the `Deploy` or `Execute` message execution, or until they are consumed.

An alternative would have been to add a `funds` field, like it happens in cosmwasm, which guarantees the called contract that
the funds are available and sent in the context of the message execution. This would have been a simpler approach, but it would
have been limited to the context of `MsgSend` only, where the asset is `sdk.Coins`. The proposed generic way, instead, allows
the account to execute any message on behalf of the sender, which is more flexible, it could include NFT send execution, or 
more complex things like `MsgMultiSend` or `MsgDelegate`, etc.


### Further discussion

#### Sub-accounts

We could provide a way to link accounts to other accounts. Maybe during deployment the sender could decide to link the 
newly created to its own account, although there might be use-cases for which the deployer is different from the account
that needs to be linked, in this case a handshake protocol on linking would need to be defined.

#### Predictable address creation

We need to provide a way to create an account with a predictable address, this might serve a lot of purposes, like accounts
wanting to generate an address that:
- nobody else can claim besides the account used to generate the new account
- is predictable

For example:
```protobuf

message MsgDeployPredictable {
  string sender = 1;
  uint32 nonce = 2; 
  ...
}
```

And then the address becomes `bechify(concat(sender, nonce))`

`x/accounts` would still use the monotonically increasing sequence as account number.

## x/authn and account credentials

We have initially mentioned that `x/accounts` only provides for a way to deploy accounts, but they're not backed by any 
credential, which means no state transition on behalf of the account can be incepted from a TX.

`x/authn` is meant to provide the link between accounts and their TX credentials.
Specifically `x/authn` defines the following `tx` gRPC interface:

```protobuf
// Msg defines the Msg service.
service Msg {
  rpc CreateAuthenticatedAccount(MsgCreateAuthenticatedAccount) returns (MsgCreateAuthenticatedAccountResponse) {};
  rpc UpdateCredentials(MsgUpdateCredentials) returns (MsgUpdateCredentialsResponse) {};
  rpc DeleteCredentials(MsgDeleteCredentials) returns (MsgDeleteCredentials) {};
}

message MsgCreateAuthenticatedAccount {
  string sender = 1;
  google.Protobuf.Any credential = 2;
  accounts.MsgDeploy deploy_msg = 3;
}

message MsgCreateAuthenticatedAccountResponse {
  accounts.MsgDeployResponse deploy_response = 1;
}

message MsgUpdateCredentials {
  string sender = 1;
  string kind = 2;
  google.Protobuf.Any  new_credential = 3;
}

message MsgUpdateCredentialsResponse {}

message MsgDeleteCredentials {
  string sender = 1;
}
message MsgDeleteCredentialsResponse {}
```

### MsgCreateAuthenticatedAccount

This message contains an opaque credential defined as `google.Protobuf.Any`, alongside an `x/accounts/MsgDeploy` request.
This creates a new account and couples it with a credential.

### MsgUpdateCredentials & MsgDeleteCredentials

The former allows for credentials of an account to be updated the latter destroys credentials for an account, making it 
effectively impossible for the account to send state transitions from a TX forever, unless the account has logic to again
update its credentials.

### The credential interface

The credential interface, represented in our gRPC `tx` interface as a `google.Protobuf.Any`, is implemented by any type 
which satisfies the following interface:

```go
package authn

type Credential[T any, PT interface{ *T; proto.Message }] interface {
	*T
	VerifySignedBytes(msgBytes []byte, signature []byte) bool
}
```
#### VerifySignedBytes

The credential is fetched from state based on the entity trying to authenticate the `state transition`, we know this by
pulling the signer of the message from the message itself (currently). Then the credential just applies the verification
logic.

### Further discussion

#### Credentials abstraction 

Credentials are currently abstracted over entities that can verify arbitrary signed bytes, which covers the use-case of crypto
curves.

The idea is that in the future we can further abstract the authentication mechanism (not only the curve), this means that a chain
can be able to define its own authentication mechanisms, which are not tied to the `SignMode` provided by the sdk.
Trying to fit this change right now would have yielded into a much broader work, that would have most likely lead to impactful
breaking changes.

#### Update x/auth, instead of creating an x/authn module

Considering the limitations over the credentials abstraction, the changeset for `x/auth` is still limited, so we could
just update `x/auth` to support `x/accounts` deployment and key rotation messages.

## Migration

### Phase1

- Implement `x/accounts` and `x/authn`
- Accounts can be migrated explicitly either by sending `x/auth` a `MsgMigrateToAuthn`, or implicitly in the `AnteHandler`.
- Move `auth` permissions  from `x/auth` to `x/bank`, as bank is the only consumer of this information.

## Major implications

- Account unique identifier (represented as AccAddress) is decoupled from the authentication mechanism.
- Accounts need to always be explicitly created. Bank would not create an account in case it doesn't exist during a `MsgSend` execution.

