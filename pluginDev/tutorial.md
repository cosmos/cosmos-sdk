
# Basecoin Plugin Development Tutorial

Basecoin is a demonstration environment of a Tendermint powered blockchain
crypto-currency.  Within basecoin, utility can be extended through the use of
custom plugins, which extend the type of transactions from simple coin exchange
to anything your little heart can dream up, so long as its deterministic and terminating.
The typical form of a plugin transaction includes sending coins
with an associated account to the plugin along with custom transaction
information, as well as fees and gas required per transaction. 
See [here][0] for more details on the plugin design.
This tutorial will walk you through of how to write a basic plugin for basecoin and deploy your own
instance of basecoin running your plugin.  Note that this tutorial assumes that
you are familiar with golang. For more information on golang see [The Go
Blog][1] and [Go Basics][2].  A completed copy of the example code described in
this tutorial can be found [here][3].  [0]:
https://github.com/tendermint/basecoin/blob/develop/Plugins.md [1]:
https://blog.golang.org/ [2]:
https://github.com/tendermint/basecoin/blob/master/GoBasics.md [3]:
https://github.com/tendermint/basecoin-examples/tree/master/pluginDev

### Programming a custom basecoin application

First create a new project with an empty main package under your github
directory.  Within the main package add the following imports to the import
block. 

```golang "flag" "github.com/tendermint/abci/server"
"github.com/tendermint/basecoin/app" eyes
"github.com/tendermint/merkleeyes/client" cmn "github.com/tendermint/go-common"
```

Begin the main function by defining the following flags for example use within
initialization of the basecoin app for starting the merkleeyes client, loading
genesis, and starting the abci listener. 

```golang eyesPtr := flag.String("eyes", "local", "MerkleEyes address, or
'local' for embedded") genFilePath := flag.String("genesis", "", "Genesis file,
if any") addrPtr := flag.String("address", "tcp://0.0.0.0:46658", "Listen
address") flag.Parse() ```

Now we can initialize a new instance of basecoin.
First connect to a new [merkleeyes][4] client, and then initialize 
basecoin with it.  Merkleeyes is an abci utility which serves data from a
merkle-tree key-value store.  The merkle-tree instance that is loaded into
basecoin is used as a [common data store](https://github.com/tendermint/basecoin/blob/master/types/kvstore.go#L10-L13)
for the basecoin application and all of its plugins.
Having a central data store creates the potential for key collisions making it
essential to use explicit key name spaces for each plugin.  [4]:
https://github.com/tendermint/merkleeyes

```golang
// Connect to MerkleEyes
eyesCli, err := eyes.NewClient(*eyesPtr, "socket") if err != nil {
cmn.Exit("connect to MerkleEyes: " + err.Error()) }

// Create Basecoin
bcApp := app.NewBasecoin(eyesCli) ```

Next, set the initial state of the basecoin app load the genesis file (if any)
for basecoin.  

```golang
// Connect to MerkleEyes
eyesCli, err := eyes.NewClient(*eyesPtr, "socket") if err != nil {
cmn.Exit("connect to MerkleEyes: " + err.Error()) }

// Create Basecoin
bcApp := app.NewBasecoin(eyesCli)

// If genesis file was specified, set key-value options
if *genFilePath != "" { err := bcApp.LoadGenesis(*genFilePath) if err != nil {
cmn.Exit(cmn.Fmt("%+v", err)) } } ```

Finally implement an [abci](https://github.com/tendermint/abci) listening
server for the basecoin app. The listening server maintains a server connection
to tendermint for blockchain communications between basecoin nodes. 

```golang c// Start the listener svr, err := server.NewServer(*addrPtr,
"socket", bcApp) if err != nil { cmn.Exit("create listener: " + err.Error()) }

// Wait forever
cmn.TrapSignal(func() {
	// Cleanup
	svr.Stop() }) ```

Great! Your main.go file should look [like this][5] and should be able to
install and run, although the plugin has not yet been created.  [5]:
https://github.com/tendermint/basecoin-examples/blob/master/pluginDev/interim/main-noplugin.go

### Programming a basic plugin
 
The next step is to create your plugin that can be run your custom instance of
basecoin. The plugin package should be located in plugins/YourPlugin from the
repo root. In this instance we will start our plugin from this [boilerplate][7]
file by copying it to the directory`plugins/counter/counter.go`, and renaming
the package _Counter_. By definition any basecoin plugin needs to satisfy the
basecoin [Plugin interface][6]. The following is a description of the required
functions: [6]:
https://github.com/tendermint/basecoin/blob/master/types/plugin.go#L8-L15 [7]:
https://github.com/tendermint/basecoin-examples/blob/master/pluginDev/interim/plugin-blank.go

 - Name() string
   - returns the name of the application
 - SetOption(store KVStore, key string, value string) (log string)
   - SetOption may be called during genesis of basecoin and can be used to set
     initial plugin parameters 
 - RunTx(store KVStore, ctx CallContext, txBytes []byte) (res abci.Result)
   - RunTx contains the core logic of the plugin. This function is called
     during operation a plugins transaction is called 
 - InitChain(store KVStore, vals []\*abci.Validator)
   - Not-used within this tutorial
 - BeginBlock(store KVStore, height uint64)
   - Not-used within this tutorial
 - EndBlock(store KVStore, height uint64) []\*abci.Validator
   - Not-used within this tutorial

The boilerplate plugin file provided also includes a non-required function
named _StateKey_.  This function is used to define a unique keyspace specific
to the plugins name that can be used to set and retrieve values from the
KVStore.  

Three plugin structs are provided the boilerplate file, each serving a unique
function for the plugins operation.   
  
1. Plugin 
  * Plugin struct which satisfies the [Plugin interface][6] as defined by
  * basecoin Stores the name of the plugin instance Initialized using the _New_
  * func in the plugin-blank.go To be rename CounterPlugin in this tutorial
  * resulting struct for this tutorial:   
    
    ```golang  type CounterPlugin struct {  name string  }  ```  

2. PluginTx 
  * Used by [go-wire][8] as the encoding/decoding struct to pass transaction
  * data through txBytes in the _RunTx_ func. 
[8]: https://github.com/tendermint/go-wire
  * Stores transaction-specific plugin-defined information to be renamed
  * CounterTx in this tutorial and and include two custom transaction variables
  * Valid: should the transaction be performed Fee: cost that the CounterTx
  * Plugin charges per interaction resulting struct for this tutorial: 
   
    ```golang type CounterTx struct { Valid bool Fee  types.Coins } ```

3. PluginState
  * Used by go-wire as the encoding/decoding struct to hold the plugin state
  * data in the KVStore used within the _RunTx_ func.  Stores plugin-specific
  * information to be renamed CounterPluginState in this tutorial and include
  * two custom state variables Counter: count for the number of transactions
  * TotalFees: the total amount paid to the counter plugin through transactions 
    ```golang type CounterPluginState struct { Counter   int TotalFees
types.Coins } ```


The core tx logic of the app is containted within the RunTx function. RunTx
contains three input fields:
 - store types.KVStore
   - This term provides read/write capabilities to the merkelized data store
     which is accessible cross-plugin
 - ctx types.CallContext
   - The ctx contains the callers address, a pointer to the callers account,
     and an amount of coins sent with the transaction
 - txBytes []byte
   - This field can be used to send customized information from the basecoin
     application to your plugin

There are five basic steps in performing the app logic for a simple plugin such
as counter. Within the RunTx function:  1. Decode the transaction bytes
(txBytes)  2. Validate the transaction   3. Load and decode the app state from
the KVStore  4. Perform plugin transaction logic  5. Update and save the app
state to the KVStore  

Other more complex plugins may have a variant on the above logic including
loading and saving multiple or variable states, or not including a state stored
in the KVStore whatsoever. 

As in the first step, decode the information from this term with [go-wire][8].
In this step we are attempting to write the txBytes to the variable tx or the
type struct CounterTx, if the txBytes have not been properly encoded from a
CounterTx struct wire will produce an error.

```golang
// Decode tx
var tx CounterTx err := wire.ReadBinaryBytes(txBytes, &tx) if err != nil {
return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
} ```

Step two is to perform some checks to validate the transaction. First check the
_Valid_ term and terminate the process if it's set to false. Next check is the
cost term is valid and non-negative as well as if the fee has been adequately
covered by the coins sent within the transaction context (_ctx.Coins_).

```golang
// Validate tx
if !tx.Valid { return abci.ErrInternalError.AppendLog("CounterTx.Valid must be
true") }

if !tx.Cost.IsValid() { return abci.ErrInternalError.AppendLog("CounterTx.Cost
is not sorted or has zero amounts") } if !tx.Cost.IsNonnegative() { return
abci.ErrInternalError.AppendLog("CounterTx.Cost must be nonnegative") }

// Did the caller provide enough coins?
if !ctx.Coins.IsGTE(tx.Fee) { return
abci.ErrInsufficientFunds.AppendLog("CounterTx.Cost was not provided") } ```

If all the checks have passed, the core transaction logic is calculated and
saved with basecoin. The results is saved as to a new 'state' struct, encoded
to a byte array using go-wire, and saved to basecoin using the KVStore term
passed to RunTx through _store_ term. To accomplish this first load the state
if exists as provided in the boilerplate example code. Note that if the state
does not exist, the state bytes read into _cpStateBytes_ will have no members
therefore returning a length of zero. If this is the case we will skip loading
the initial state  

```golang
// Load CounterPluginState
var cpState CounterPluginState cpStateBytes := store.Get(cp.StateKey()) if
len(cpStateBytes) > 0 { err = wire.ReadBinaryBytes(cpStateBytes, &cpState) if
err != nil { return abci.ErrInternalError.AppendLog("Error decoding state: " +
err.Error()) } } ```

Now we can perform our counter logic of the state variables to reflect the
ongoing transaction, and save the transaction results back to the store. Here
we will increase the state variable counter by one, and increase the TotalFees
by the current Fee being charged by the transaction.

```golang	
// Update CounterPluginState
cpState.Counter += 1 cpState.TotalFees = cpState.TotalCost.Plus(tx.Fee)

// Save CounterPluginState
store.Set(cp.StateKey(), wire.BinaryBytes(cpState)) ```

Great! The plugin package is now complete, it should look a bit like [this][9].
It's now possible to load the counter plugin into the instance of basecoin
opened in main.go. Define and register your plugin with the basecoin instance,
after you have defined the instance but before you have run the Tendermint
listening server.  [9]:
https://github.com/tendermint/basecoin-examples/blob/master/pluginDev/completed/plugins/counter/counter.go

```golang
// create/add plugins
counter := counter.New("counter") bcApp.RegisterPlugin(counter) ```

The counter plugin package must be imported add the following to the main.go
import block, where _username_ is replaced with your username, and _reponame_
is replaced by the name of you repo.

```golang	"github.com/username/reponame/plugins/counter" ```

main.go should now look like [this][9]. A completed copy of the source code
described in this tutorial can be found [here][1] [9]:
https://github.com/tendermint/basecoin-examples/blob/master/pluginDev/completed/main.go

### Initializing Dependencies

The first step to running your application is to install your dependencies. For
golang this can be done with [glide][7] [7]:
https://github.com/Masterminds/glide
 - retrieve glide `go get github.com/Masterminds/glide`
 - initialize glide from your application's directory with `glide init`
 - follow the instructions (for quick installation choose no on first prompt)
 - install the dependencies with `glide install`

### Installing and running your application

Your application can be installed using the command `go install` when navigated
to the root directory of your application folder.  To run your new application
use the following steps:

1. In a terminal window run `counter` 2. Within a second terminal window run
`tendermint node`

Your blockchain basecoin application is now operating and can be communicated
with through the basecoin CLI!

### Developing a Custom Plugin Command Line Interface

Coming soon!
