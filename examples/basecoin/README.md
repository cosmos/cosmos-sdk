# Basecoin

This is the "Basecoin" example application built on the Cosmos-Sdk.  This
"Basecoin" is not affiliated with [Coinbase](http://www.getbasecoin.com/), nor
the [stable coin](http://www.getbasecoin.com/).

Assuming you've run `make get_tools && make get_vendor_deps` from the root of 
this repository, run `make build` here to build the `basecoind` and `basecli` 
binaries.

If you want to create a new application, start by copying the Basecoin app.


# Building your own Blockchain

Basecoin is the equivalent of an ERC20 token contract for blockchains. In order
to deploy your own application all you need to do is clone `examples/basecoin`
and run it. Now you are already running your own blockchain. In the following
I will explain how to add functionality to your blockchain. This is akin to
defining your own vesting schedule within a contract or setting a specific
multisig. You are just extending the base layer with extra functionality here
and there.

## Structure of Basecoin

Basecoin is build with the cosmos-sdk. It is a sample application that works
with any engine that implements the ABCI protocol. Basecoin defines multiple
unique modules as well as uses modules directly from the sdk. If you want
to modify Basecoin, you either remove or add modules according to your wishes.


## Modules

A module is a fundamental unit in the cosmos-sdk. A module defines its own
transaction, handles its own state as well as its own state transition logic.
Globally, in the `app/app.go` file you just have to define a key for that 
module to access some parts of the state, as well as initialise the module
object and finally add it to the transaction router. The router ensures that
every module only gets its own messages. 


## Transactions

A user can send a transaction to the running blockchain application. This
transaction can be of any of the ones that are supported by any of the 
registered modules. 

### CheckTx

Once a user has submitted their transaction to the engine,
the engine will first run `checkTx` to confirm that it is a valid transaction.
The module has to define a handler that knows how to handle every transaction
type. The corresponding handler gets invoked with the checkTx flag set to true.
This means that the handler shouldn't do any expensive operations, but it can
and should write to the checkTx state.

### DeliverTx

The engine calls `deliverTx` when a new block has been agreed upon in 
consensus. Again, the corresponding module will have its handler invoked 
and the state and context is passed in. During deliverTx execution the 
transaction needs to be processed fully and the results are written to the 
application state.


## CLI

The cosmos-sdk contains a number of helper libraries in `clients/` to build cli
and RPC interfaces for your specific application.

