# x/accounts

The x/accounts module provides module and facilities for writing smart cosmos-sdk accounts.

# The `Authentication` Interface, enabling abstracted TX authentication mechanisms beyond public key cryptography.

The x/accounts module allows chain developers to implement custom authentication mechanisms for their accounts, specifically
an account can be authenticated from a TX if it implements the `Authentication` interface, interface implementation in x/accounts' 
accounts is done by exposing an execution handler capable of handling a specific message.

In this case the message we want to implement is `MsgAuthenticate` as defined [here](./proto/cosmos/accounts/interfaces/account_abstraction/v1/interface.proto).

## How to implement the interface

An account can implement the interface by handling the execution of `MsgAuthenticate`:

```go
package base

// Authenticate implements the authentication flow of an abstracted base account.
func (a Account) Authenticate(ctx context.Context, msg *aa_interface_v1.MsgAuthenticate) (*aa_interface_v1.MsgAuthenticateResponse, error) {
	if !accountstd.SenderIsAccountsModule(ctx) {
		return nil, errors.New("unauthorized: only accounts module is allowed to call this")
	}
	// your authentication logic.
}

// Implementing the authentication interface is done by registering the execution handler 
// which handles authenticate.
func (a Account) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, a.SwapPubKey) // other handler
	accountstd.RegisterExecuteHandler(builder, a.Authenticate) // <= implements the Authentication interface.
}

```

It is important to:
1. Verify the sender is the x/accounts module. This will make it impossible for other accounts to try to trigger authentication.
2. Ensure that the authentication mechanism is safe:
   3. It must not be possible to replay the same action using the same signature.


# Supporting Custom Accounts in the x/auth gRPC Server

## Overview

The x/auth module provides a mechanism for custom account types to be exposed via its `Account` and `AccountInfo` gRPC
queries. This feature is particularly useful for ensuring compatibility with existing wallets that have not yet integrated 
with x/accounts but still need to parse account information post-migration.

## Implementation

To support this feature, your custom account type needs to implement the `auth.QueryLegacyAccount` handler. Here are some important points to consider:

1. **Selective Implementation**: This implementation is not required for every account type. It's only necessary for accounts you want to expose through the x/auth gRPC `Account` and `AccountInfo` methods.
2. **Flexible Response**: The `info` field in the `QueryLegacyAccountResponse` is optional. If your custom account cannot be represented as a `BaseAccount`, you can leave this field empty.

## Example Implementation

A concrete example of implementation can be found in `defaults/base/account.go`. Here's a simplified version:

```go
func (a Account) AuthRetroCompatibility(ctx context.Context, _ *authtypes.QueryLegacyAccount) (*authtypes.QueryLegacyAccountResponse, error) {
    seq := a.GetSequence()
    num := a.GetNumber()
    address := a.GetAddress()
    pubKey := a.GetPubKey()

    baseAccount := &authtypes.BaseAccount{
        AccountNumber: num,
        Sequence:      seq,
        Address:       address,
    }

    // Convert pubKey to Any type
    pubKeyAny, err := gogotypes.NewAnyWithValue(pubKey)
    if err != nil {
        return nil, err
    }
    baseAccount.PubKey = pubKeyAny

    // Convert the entire baseAccount to Any type
    accountAny, err := gogotypes.NewAnyWithValue(baseAccount)
    if err != nil {
        return nil, err
    }

    return &authtypes.QueryLegacyAccountResponse{
        Account: accountAny,
        Info:    baseAccount,
    }, nil
}
```

## Usage Notes

- Implement this handler only for account types you want to expose via x/auth gRPC methods.
- The `info` field in the response can be nil if your account doesn't fit the `BaseAccount` structure.

# Genesis

## Creating accounts on genesis

In order to create accounts at genesis, the `x/accounts` module allows developers to provide
a list of genesis `MsgInit` messages that will be executed in the `x/accounts` genesis flow.

The init messages are generated offline. You can also use the following CLI command to generate the
json messages: `simd accounts tx init [account type] [msg] --from me --genesis`. This will generate 
a jsonified init message wrapped in an x/accounts `MsgInit`.

This follows the same initialization flow and rules that would happen if the chain is running. 
The only concrete difference is that this is happening at the genesis block.

For example, given the following `genesis.json` file:

```json
{
  "app_state": {
    "accounts": {
      "init_account_msgs": [
        {
          "sender": "account_creator_address",
          "account_type": "lockup",
          "message": {
            "@type": "cosmos.accounts.defaults.lockup.MsgInitLockupAccount",
            "owner": "some_owner",
            "end_time": "..",
            "start_time": ".."
          },
          "funds": [
            {
              "denom": "stake",
              "amount": "1000"
            }
          ]
        }
      ]
    }
  }
}
```

The accounts module will run the lockup account initialization message.