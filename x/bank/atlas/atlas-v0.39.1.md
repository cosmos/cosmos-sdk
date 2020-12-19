# x/bank

The `x/bank` module is responsible for handling multi-asset coin transfers between
accounts and tracking special-case pseudo-transfers which must work differently
with particular kinds of accounts.

## Usage

1. Import the module.

   ```go
   import (
       "github.com/cosmos/cosmos-sdk/x/bank"
   )
   ```

2. Add `AppModuleBasic` to your `ModuleBasics`.

    ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        bank.AppModuleBasic{},
      }
    )
    ```

3. Create the module's parameter subspace in your application constructor.

   ```go
   func NewApp(...) *App {
     // ...
     app.subspaces[bank.ModuleName] = app.ParamsKeeper.Subspace(bank.DefaultParamspace)
   }
   ```

4. Create the keeper. Note, the `x/bank` module depends on the `x/auth` module
   and a list of blacklisted account addresses which funds are not allowed to be
   sent to. Your application will need to define this method based your needs.

   ```go
   func NewApp(...) *App {
     // ...
     app.BankKeeper = bank.NewBaseKeeper(
       app.AccountKeeper, app.subspaces[bank.ModuleName], app.BlacklistedAccAddrs(),
     )
   }
   ```

5. Add the `x/bank` module to the app's `ModuleManager`.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm = module.NewManager(
       // ...
       bank.NewAppModule(app.BankKeeper, app.AccountKeeper),
       // ...
     )
   }
   ```

6. Set the `x/bank` module genesis order.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm.SetOrderInitGenesis(..., bank.ModuleName, ...)
   }
   ```

7. Add the `x/bank` module to the simulation manager (if you have one set).

   ```go
   func NewApp(...) *App {
     // ...
     app.sm = module.NewSimulationManager(
       // ...
       bank.NewAppModule(app.BankKeeper, app.AccountKeeper),
       // ...
     )
   }

## Genesis

The `x/bank` module defines its genesis state as follows:

```go
type GenesisState struct {
  SendEnabled bool `json:"send_enabled" yaml:"send_enabled"`
}
```

The `SendEnabled` parameter determines if transfers are enabled or disabled
entirely on the chain. This can be used to start a network without enabling
transfers while ensuring critical network functionality is operating as expected.

## Messages

### `MsgSend`

The `x/bank` module allows for transfer of funds from a source account to a
destination account.

```go
type MsgSend struct {
  FromAddress sdk.AccAddress `json:"from_address" yaml:"from_address"`
  ToAddress   sdk.AccAddress `json:"to_address" yaml:"to_address"`
  Amount      sdk.Coins      `json:"amount" yaml:"amount"`
}
```

### `MsgMultiSend`

The `x/bank` module also allows for multiple inputs and outputs. The sum of all
inputs must be equivalent to the sum of all outputs.

```go
type Input struct {
  Address sdk.AccAddress `json:"address" yaml:"address"`
  Coins   sdk.Coins      `json:"coins" yaml:"coins"`
}

type Output struct {
  Address sdk.AccAddress `json:"address" yaml:"address"`
  Coins   sdk.Coins      `json:"coins" yaml:"coins"`
}

type MsgMultiSend struct {
  Inputs  []Input  `json:"inputs" yaml:"inputs"`
  Outputs []Output `json:"outputs" yaml:"outputs"`
}
```

## Client

### CLI

The `x/bank` supports the following transactional commands.

1. Send tokens via a `MsgSend` message.

   ```shell
   $ app tx send [from_key_or_address] [to_address] [amount] [...flags]
   ```

Note, the `x/bank` module does not natively support constructing a `MsgMultiSend`
message. This type of message must be constructed manually, but it may be signed
and broadcasted via the CLI.

### REST

The `x/bank` supports various query API endpoints and a `MsgSend` construction
endpoint.

1. Construct an unsigned `MsgSend` transaction.

   | Method | Path                     |
   | :----- | :----------------------- |
   | `POST` | `/bank/accounts/{address}/transfers` |

   Sample payload:

   ```json
   {
       "base_req": {
           "chain_id": "chain-foo",
           "from": "cosmos1u3fneykx9carelvurc6av22vpjvptytj9wklk0",
           "memo": "memo",
           "fees": [
               {
                   "denom": "stake",
                   "amount": "25000"
               }
           ]
       },
       "amount": [
           {
               "denom": "stake",
               "amount": "400000000"
           }
       ]
   }
   ```

2. Query for an account's balance.

   | Method | Path                     |
   | :----- | :----------------------- |
   | `GET` | `/bank/balances/{address}` |

   Sample response:

   ```json
   {
       "height": "0",
       "result": [
           {
               "denom": "node0token",
               "amount": "1000000000"
           },
           {
               "denom": "stake",
               "amount": "400000000"
           }
       ]
   }
   ```
