Basecoin Extensions
===================

TODO: re-write for extensions

In the `previous guide <basecoin-basics.md>`__, we saw how to use the
``basecoin`` tool to start a blockchain and the ``basecli`` tools to
send transactions. We also learned about ``Account`` and ``SendTx``, the
basic data types giving us a multi-asset cryptocurrency. Here, we will
demonstrate how to extend the tools to use another transaction type, the
``AppTx``, so we can send data to a custom plugin. In this example we
explore a simple plugin named ``counter``.

Example Plugin
--------------

The design of the ``basecoin`` tool makes it easy to extend for custom
functionality. The Counter plugin is bundled with basecoin, so if you
have already `installed basecoin <install.md>`__ and run
``make install`` then you should be able to run a full node with
``counter`` and the a light-client ``countercli`` from terminal. The
Counter plugin is just like the ``basecoin`` tool. They both use the
same library of commands, including one for signing and broadcasting
``SendTx``.

Counter transactions take two custom inputs, a boolean argument named
``valid``, and a coin amount named ``countfee``. The transaction is only
accepted if both ``valid`` is set to true and the transaction input
coins is greater than ``countfee`` that the user provides.

A new blockchain can be initialized and started just like in the
`previous guide <basecoin-basics.md>`__:

::

    # WARNING: this wipes out data - but counter is only for demos...
    rm -rf ~/.counter
    countercli reset_all

    countercli keys new cool
    countercli keys new friend

    counter init $(countercli keys get cool | awk '{print $2}')

    counter start

The default files are stored in ``~/.counter``. In another window we can
initialize the light-client and send a transaction:

::

    countercli init --node=tcp://localhost:26657 --genesis=$HOME/.counter/genesis.json

    YOU=$(countercli keys get friend | awk '{print $2}')
    countercli tx send --name=cool --amount=1000mycoin --to=$YOU --sequence=1

But the Counter has an additional command, ``countercli tx counter``,
which crafts an ``AppTx`` specifically for this plugin:

::

    countercli tx counter --name cool
    countercli tx counter --name cool --valid

The first transaction is rejected by the plugin because it was not
marked as valid, while the second transaction passes. We can build
plugins that take many arguments of different types, and easily extend
the tool to accomodate them. Of course, we can also expose queries on
our plugin:

::

    countercli query counter

Tada! We can now see that our custom counter plugin transactions went
through. You should see a Counter value of 1 representing the number of
valid transactions. If we send another transaction, and then query
again, we will see the value increment. Note that we need the sequence
number here to send the coins (it didn't increment when we just pinged
the counter)

::

    countercli tx counter --name cool --countfee=2mycoin --sequence=2 --valid
    countercli query counter

The Counter value should be 2, because we sent a second valid
transaction. And this time, since we sent a countfee (which must be less
than or equal to the total amount sent with the tx), it stores the
``TotalFees`` on the counter as well.

Keep it mind that, just like with ``basecli``, the ``countercli``
verifies a proof that the query response is correct and up-to-date.

Now, before we implement our own plugin and tooling, it helps to
understand the ``AppTx`` and the design of the plugin system.

AppTx
-----

The ``AppTx`` is similar to the ``SendTx``, but instead of sending coins
from inputs to outputs, it sends coins from one input to a plugin, and
can also send some data.

::

    type AppTx struct {
      Gas   int64   `json:"gas"`
      Fee   Coin    `json:"fee"`
      Input TxInput `json:"input"`
      Name  string  `json:"type"`  // Name of the plugin
      Data  []byte  `json:"data"`  // Data for the plugin to process
    }

The ``AppTx`` enables Basecoin to be extended with arbitrary additional
functionality through the use of plugins. The ``Name`` field in the
``AppTx`` refers to the particular plugin which should process the
transaction, and the ``Data`` field of the ``AppTx`` is the data to be
forwarded to the plugin for processing.

Note the ``AppTx`` also has a ``Gas`` and ``Fee``, with the same meaning
as for the ``SendTx``. It also includes a single ``TxInput``, which
specifies the sender of the transaction, and some coins that can be
forwarded to the plugin as well.

Plugins
-------

A plugin is simply a Go package that implements the ``Plugin``
interface:

::

    type Plugin interface {

      // Name of this plugin, should be short.
      Name() string

      // Run a transaction from ABCI DeliverTx
      RunTx(store KVStore, ctx CallContext, txBytes []byte) (res abci.Result)

      // Other ABCI message handlers
      SetOption(store KVStore, key string, value string) (log string)
      InitChain(store KVStore, vals []*abci.Validator)
      BeginBlock(store KVStore, hash []byte, header *abci.Header)
      EndBlock(store KVStore, height uint64) (res abci.ResponseEndBlock)
    }

    type CallContext struct {
      CallerAddress []byte   // Caller's Address (hash of PubKey)
      CallerAccount *Account // Caller's Account, w/ fee & TxInputs deducted
      Coins         Coins    // The coins that the caller wishes to spend, excluding fees
    }

The workhorse of the plugin is ``RunTx``, which is called when an
``AppTx`` is processed. The ``Data`` from the ``AppTx`` is passed in as
the ``txBytes``, while the ``Input`` from the ``AppTx`` is used to
populate the ``CallContext``.

Note that ``RunTx`` also takes a ``KVStore`` - this is an abstraction
for the underlying Merkle tree which stores the account data. By passing
this to the plugin, we enable plugins to update accounts in the Basecoin
state directly, and also to store arbitrary other information in the
state. In this way, the functionality and state of a Basecoin-derived
cryptocurrency can be greatly extended. One could imagine going so far
as to implement the Ethereum Virtual Machine as a plugin!

For details on how to initialize the state using ``SetOption``, see the
`guide to using the basecoin tool <basecoin-tool.md#genesis>`__.

Implement your own
------------------

To implement your own plugin and tooling, make a copy of
``docs/guide/counter``, and modify the code accordingly. Here, we will
briefly describe the design and the changes to be made, but see the code
for more details.

First is the ``cmd/counter/main.go``, which drives the program. It can
be left alone, but you should change any occurrences of ``counter`` to
whatever your plugin tool is going to be called. You must also register
your plugin(s) with the basecoin app with ``RegisterStartPlugin``.

The light-client is located in ``cmd/countercli/main.go`` and allows for
transaction and query commands. This file can also be left mostly alone
besides replacing the application name and adding references to new
plugin commands.

Next is the custom commands in ``cmd/countercli/commands/``. These files
are where we extend the tool with any new commands and flags we need to
send transactions or queries to our plugin. You define custom ``tx`` and
``query`` subcommands, which are registered in ``main.go`` (avoiding
``init()`` auto-registration, for less magic and more control in the
main executable).

Finally is ``plugins/counter/counter.go``, where we provide an
implementation of the ``Plugin`` interface. The most important part of
the implementation is the ``RunTx`` method, which determines the meaning
of the data sent along in the ``AppTx``. In our example, we define a new
transaction type, the ``CounterTx``, which we expect to be encoded in
the ``AppTx.Data``, and thus to be decoded in the ``RunTx`` method, and
used to update the plugin state.

For more examples and inspiration, see our `repository of example
plugins <https://github.com/tendermint/basecoin-examples>`__.

Conclusion
----------

In this guide, we demonstrated how to create a new plugin and how to
extend the ``basecoin`` tool to start a blockchain with the plugin
enabled and send transactions to it. In the next guide, we introduce a
`plugin for Inter Blockchain Communication <ibc.md>`__, which allows us
to publish proofs of the state of one blockchain to another, and thus to
transfer tokens and data between them.
