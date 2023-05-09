# RFC 005: Authentication

## Changelog

- 09/05/2023: DRAFT

## Problem Statement

The recently proposed changes to the account system in the Cosmos SDK (as described in the Accounts RFC) focus on
transforming accounts into more sophisticated, module-like entities capable of managing their own state. However, several
key issues need to be addressed with respect to authentication:

### 1. Account Abstraction

The Accounts RFC provides an abstraction over accounts as business logic entities, similar to modules, but does not
address how authentication should be handled for these new account types.

### 2. Accounts without Credentials

It is possible for accounts to exist without any associated credentials, which means they cannot initiate state 
transitions through transactions. A mechanism is needed to attach authentication credentials to such accounts to enable
them to interact with the network.

### 3. Pluggable Authentication Credentials

The authentication system should provide a way to plug different credential mechanisms into accounts. This would allow 
developers to customize the authentication process for different account types or use cases.

### 4. Extensible Credential Mechanisms
   
Currently, the Cosmos SDK primarily supports accounts backed by public/private key pairings. However, there is a growing
need to design and implement an extensible system that can support abstracted credentials, which are not limited to
public/private key pairings. This would enable the development of more advanced and flexible authentication systems that
cater to a broader range of use cases and security requirements.

To address these problems, the Cosmos SDK needs to develop an authentication system that is closely integrated with the
proposed account system, providing a robust and extensible framework for managing authentication credentials and supporting
a variety of credential mechanisms.

## Implementation

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
  google.Protobuf.Any new_credential = 3;
}
message MsgUpdateCredentialsResponse {}
message MsgDeleteCredentials {
  string sender = 1;
}
message MsgDeleteCredentialsResponse {}
```

### MsgCreateAuthenticatedAccount

This message contains an opaque credential defined as `google.Protobuf.Any`, alongside an `accounts.MsgDeploy` request.
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
