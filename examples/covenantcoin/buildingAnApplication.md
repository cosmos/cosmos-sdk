# Building an Application on the Cosmos SDK
Note this is written from my understanding of writing an application, not all information may be correct.
TODO: Add CLI stuff

## Description of the application (Covenant Coin)

Suppose I want to make an application to have covenants. A covenant stores money, and it can only be paid out to certain addresses. When a covenant is created, you have to specify who is allowed to pay the money from the covenant. For simplicity, we are going to make it such that when a covenant is settled, all money just goes to one address.

Before digging into how to build such an application, I want to first explain a bit about how SDK applications are structured.

## Structure
Lets look briefly at democoin
```
democoin/
├── app
│   ├── app.go
│   └── app_test.go
├── cmd
│   └── Stuff that handles the daemon / cli
├── types
│   └── account.go
└── x
    ├── simplestake
    │   ├── client (cli stuff here)
    │   ├── errors.go
    │   ├── handler.go
    │   ├── keeper.go
    │   ├── keeper_test.go
    │   ├── msgs.go
    │   ├── msgs_test.go
    │   ├── types.go
    │   └── wire.go
    └── other modules
```

`app.go` handles basically taking all the functionality specified, and turning it into a blockchain. (TODO: improve/fix this description) For our purposes, we're basically just going to copy paste all of basecoin's `app.go`.

`types/account.go` sets what is actually stored inside of an account. For our purposes, we're not going to need to change any of this. In general, put items in the account only if you need said data every single time you get the account. (Coins are in here for example, because you need them every time for fees) Everything else can be in a KVStore with the account as the key.


`x/simplestake` : `simplestake` is a module. This is basically where all of the code lives that handles staking. Similarly, we are going to have to organize all of our covenant handling code in our own `covenant` module. All of the things inside the module are basically registered with `app.go` later.

In the simplestake module, there are `msgs`, `keepers`, and `handlers`. Typically, one makes a separate key value store for each module. `Keepers` have the functionality of accessing / doing things with the key value store. For example, the bank module, which controls creation and deletion of coins, has a keeper to add/subtract coins from anyone, and a keeper to just view someones coins. Keepers from one application are often passed into other application's keepers. This will all make more sense when we walk through building the covenant application. Note, there is not an interface which keepers implement, they are just this concept of the thing that handles state access.

`handlers` on the other hand are given access to the minimum amount of keepers which they need, and they basically parse the message type into the relevant data, and pass that into a keeper. They then return the result.

`msgs` are the new message types you need for the application.

If the above is confusing / unclear, don't worry (it was for us also), it will make more sense after starting the application.

## Creating the application
First we're just going to fork the sdk, and copy basecoin into a new folder in `examples/covenantcoin`. (You could alternatively just create your own repository, and copy all of basecoin) Then refactor the BasecoinApp to CovenantApp, and NewBaseApp to NewCovenantApp.

We want these Covenants to be objects that exist on chain, and can be created and settled via new types of messages on chain.

The order for building this will be making a key value storage for our covenants, defining what a Covenant object is, then defining the messages, creating the Keepers which use the Key Value Storage, and then writing the handlers.

### Setting up the Key Value Storage (KVStore)
The way accessing the key value storage works is that there is a key to to the actual KVStore (of type `*sdk.KVStoreKey`), and you use this key with the context to get the current instance of the KVStore when processing a transaction. So in our `CovenantApp` struct we just add the line `keyCovenant *sdk.KVStoreKey`. Similarly, in the NewCovenantApp function we add `keyCovenant: sdk.NewKVStoreKey("covenant"),` in order to actually set this key.

Then in the line for `app.MountStoresIAVL`, add `app.keyCovenant` as a parameter. This is what actually creates the KVStore at that key.

### Creating a Covenant Object
We first need to create a folder for the covenant module `/covenantcoin/x/covenant/` and in there add `types.go`. Our covenant needs to know who can sign off on releasing the money, who the money can go to, and the amount of coins / what coins are stored in it.

This makes our `types.go`:
```
package covenant

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Covenant struct {
	Settlers  []sdk.Address
	Receivers []sdk.Address
	Amount    sdk.Coins
}
```
These are going to be stored inside of our KVStore by a global covenant ID, that gets incremented every time a covenant is created.

### Defining the messages
This will go in a `msgs.go` file inside of the module. There are two different types of messages we need, a message to create a covenant, and a message to settle a covenant. This is done with the following:
```
type MsgCreateCovenant struct {
	Sender    sdk.Address   `json:"sender"`
	Settlers  []sdk.Address `json:"settlers"`
	Receivers []sdk.Address `json:"receivers"`
	Amount    sdk.Coins     `json:"amount"`
}
type MsgSettleCovenant struct {
	CovID    int64       `json:"covid"`
	Settler  sdk.Address `json:"settler"`
	Receiver sdk.Address `json:"receiver"`
}
```
(Note that after creating a covenant successfully, the covenant id is returned)

However, we need these structs to implement the [sdk.message](https://github.com/jlandrews/cosmos-sdk/blob/master/types/tx_msg.go) interface.
* `Type()`: both of these structs will return the string `covenant`. This string returned from `Type()` is used for routing this message type to the correct handler. (It will make more sense when we get to the handler section) The handler will then switch based upon message type.
* `GetSignBytes()`: we return the struct JSON encoded, since every field needs to be signed by the relevant party.
* `ValidateBasic()`: We just check that each field in each struct is not null.
* `GetSigners()`: For MsgCreateCovenant the sender has to sign the message, for MsgSettleCovenant the Settler has to sign the message.

See the file for the actual code for the Msg struct implementations.

We now need to tell the application how to encode our Msg's. We do this by creating another file `wire.go`.

Inside of `wire.go` we do the following, to register our message types with the codec.
```
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgCreateCovenant{}, "covenant/create", nil)
	cdc.RegisterConcrete(MsgSettleCovenant{}, "covenant/settle", nil)
}
```

Then in `app.go`, we add `covenant.RegisterWire(cdc)` to register wire.

### Keepers
The keeper is where we are going to have all the functionality that involves the accessing the storage.

There are two message types we can have regarding covenants, creating a covenant, and settling a covenant. So we can have one keeper to create and settle these covenants. In the future there may be an application which should be able to view covenants, but not be able to create covenants, so you could create a second keeper with this limited access just for viewing covenants. For simplicity we're just going to write one keeper for now.

Our keeper will need access to the covenant KVStore, the bank keeper (to add/subtract funds from people), and to the codec to know how to encode data for the KVStore. Notice that we can reuse another keeper to handle the money transferral API. This looks like the following:
```
type Keeper struct {
	covStoreKey sdk.StoreKey
	bankKeeper  bank.Keeper
	cdc         *wire.Codec
}
```
We also make a straightforward `CreateKeeper` method.
After this, we create the methods we need on the keeper, such as
```
func (keeper Keeper) settleCovenant(ctx sdk.Context, covID int64,
  Settler sdk.Address, Receiver sdk.Address) sdk.Error {
    ...
  }
```
All of the actual functionality of creating and settling covenants (i.e. get's and sets from kv store, checking if the relevant acct a large enough bank balance, adding/subtracting the amount) is done inside of the keeper. See the `keeper.go` file for the details of any particular function. As a general note, keepers aren't supposed to have a concept of message types, so that other keepers can reuse the functionality without having to recreate that message type.

Since the keys and values for the kvstore are both []byte, we use the Codec passed into the keeper for marshaling the relevant key/value into bytes.

### Handlers

The handler takes our message, and the current context, and basically uses the keeper to perform the relevant actions. We can then register the handler with the sdk in the application, and it will handle everything else. The [handler type](https://github.com/cosmos/cosmos-sdk/blob/develop/types/handler.go) is `type Handler func(ctx Context, msg Msg) Result`. Since we want our handler to have knowledge of the Keeper, we return a [function closure](https://en.wikipedia.org/wiki/Closure_(computer_programming)) using the keeper. We make this handler handle both types of messages. This ends up looking like:
```
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgCreateCovenant:
			return handleMsgCreate(ctx, k, msg)
		case MsgSettleCovenant:
			return handleMsgSettle(ctx, k, msg)
		default:
			errMsg := "Unrecognized Covenant Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
}}}
```
The handling of the settling covenant is straightforward. For the msg creation covenant, since we want the result to contain the new covenant ID, we include this in the data parameter of the result.

Now all thats left is letting the sdk know that all of this exists.

To create the keeper, add it to the `CovenantApp` struct, and then add `app.covKeeper = covenant.NewKeeper(app.cdc, app.keyCovenant, app.coinKeeper)` in NewCovenantApp. Note that here is basically where we are giving the covenant keeper the ability to mint / delete coins, since we pass in that keeper.

To add the handler, in the app.Router code block, add the newly created handler, with `covenant` being the route. (This has to be the same as the route specified in the Msg's `Type()` function)

### Writing Tests

All thats left is writing some tests to show that our application works! First in both `app.go` and `app_test.go` change the import of `.../basecoin/types` to `.../covenantcoin/types`.

For our test, we initialize the blockchain to give an account `x` money. We then have them create a couple of covenants, and have then settle a covenant. We check that a covenant can't be settled multiple times, and that at the end the correct amount of money is in each account. The code shows how to make and send each of our custom message types, so questions about how to make tests should hopefully be clear after looking at the testfile. (The relevant new test is in `custom_test.go`, the default helper methods are in `app_test.go`). Note that in order to check the data field for creating a covenant, SignCheckDeliver was modified to return the result.

And we see that covenants work as expected.

### CLI Stuff
TODO: we need to add the code so that this all works from the command line.
