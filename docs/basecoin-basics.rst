.. raw:: html

   <!--- shelldown script template, see github.com/rigelrozanski/shelldown
   #!/bin/bash

   testTutorial_BasecoinBasics() {

       #shelldown[1][3] >/dev/null
       #shelldown[1][4] >/dev/null
       KEYPASS=qwertyuiop

       RES=$((echo $KEYPASS; echo $KEYPASS) | #shelldown[1][6])
       assertTrue "Line $LINENO: Expected to contain safe, got $RES" '[[ $RES == *safe* ]]'
       RES=$((echo $KEYPASS; echo $KEYPASS) | #shelldown[1][7])
       assertTrue "Line $LINENO: Expected to contain safe, got $RES" '[[ $RES == *safe* ]]'

       #shelldown[3][-1]
       assertTrue "Expected true for line $LINENO" $?

       #shelldown[4][-1] >>/dev/null 2>&1 &
       sleep 5
       PID_SERVER=$!
       disown

       RES=$((echo y) | #shelldown[5][-1] $1)
       assertTrue "Line $LINENO: Expected to contain validator, got $RES" '[[ $RES == *validator* ]]'

       #shelldown[6][0]
       #shelldown[6][1]
       RES=$(#shelldown[6][2] | jq '.data.coins[0].denom' | tr -d '"')
       assertTrue "Line $LINENO: Expected to have stringss, got $RES" '[[ $RES == strings ]]'
       RES="$(#shelldown[6][3] 2>&1)"
       assertTrue "Line $LINENO: Expected to contain ERROR, got $RES" '[[ $RES == *ERROR* ]]'

       RES=$((echo $KEYPASS) | #shelldown[7][-1] | jq '.deliver_tx.code')
       assertTrue "Line $LINENO: Expected 0 code deliver_tx, got $RES" '[[ $RES == 0 ]]'

       RES=$(#shelldown[8][-1] | jq '.data.coins[0].amount')
       assertTrue "Line $LINENO: Expected to contain 1000 strings, got $RES" '[[ $RES == 1000 ]]'

       RES=$((echo $KEYPASS) | #shelldown[9][-1] | jq '.deliver_tx.code')
       assertTrue "Line $LINENO: Expected 0 code deliver_tx, got $RES" '[[ $RES == 0 ]]'

       RES=$((echo $KEYPASS) | #shelldown[10][-1])
       assertTrue "Line $LINENO: Expected to contain insufficient funds error, got $RES" \
           '[[ $RES == *"Insufficient Funds"* ]]'

       #perform a substitution within the final tests
       HASH=$((echo $KEYPASS) | #shelldown[11][-1] | jq '.hash' | tr -d '"')
       PRESUB="#shelldown[12][-1]"
       RES=$(eval ${PRESUB/<HASH>/$HASH})
       assertTrue "Line $LINENO: Expected to not contain Error, got $RES" '[[ $RES != *Error* ]]'
   }

   oneTimeTearDown() {
       kill -9 $PID_SERVER >/dev/null 2>&1
       sleep 1
   }

   # load and run these tests with shunit2!
   DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
   . $DIR/shunit2
   -->

Basecoin Basics
===============

Here we explain how to get started with a basic Basecoin blockchain, how
to send transactions between accounts using the ``basecoin`` tool, and
what is happening under the hood.

Install
-------

With go, it's one command:

..  code:: shelldown[0]

::

    go get -u github.com/tendermint/basecoin/cmd/...

If you have trouble, see the `installation guide <./install.html>`__.

Note the above command installs two binaries: ``basecoin`` and
``basecli``. The former is the running node. The latter is a
command-line light-client. This tutorial assumes you have a 'fresh'
working environment. See how to clean up below.

Generate some keys
~~~~~~~~~~~~~~~~~~

Let's generate two keys, one to receive an initial allocation of coins,
and one to send some coins to later:

..  code:: shelldown[1]

::

    basecli keys new cool
    basecli keys new friend

You'll need to enter passwords. You can view your key names and
addresses with ``basecli keys list``, or see a particular key's address
with ``basecli keys get <NAME>``.

Initialize Basecoin
-------------------

To initialize a new Basecoin blockchain, run:

..  code:: shelldown[2]

::

    basecoin init <ADDRESS>

If you prefer not to copy-paste, you can provide the address
programatically:

..  code:: shelldown[3]

::

    basecoin init $(basecli keys get cool | awk '{print $2}')

This will create the necessary files for a Basecoin blockchain with one
validator and one account (corresponding to your key) in
``~/.basecoin``. For more options on setup, see the `guide to using the
Basecoin tool </docs/guide/basecoin-tool.md>`__.

If you like, you can manually add some more accounts to the blockchain
by generating keys and editing the ``~/.basecoin/genesis.json``.

Start Basecoin
~~~~~~~~~~~~~~

Now we can start Basecoin:

..  code:: shelldown[4]

::

    basecoin start

You should see blocks start streaming in!

Initialize Light-Client
-----------------------

Now that Basecoin is running we can initialize ``basecli``, the
light-client utility. Basecli is used for sending transactions and
querying the state. Leave Basecoin running and open a new terminal
window. Here run:

..  code:: shelldown[5]

::

    basecli init --node=tcp://localhost:46657 --genesis=$HOME/.basecoin/genesis.json

If you provide the genesis file to basecli, it can calculate the proper
chainID and validator hash. Basecli needs to get this information from
some trusted source, so all queries done with ``basecli`` can be
cryptographically proven to be correct according to a known validator
set.

Note: that ``--genesis`` only works if there have been no validator set
changes since genesis. If there are validator set changes, you need to
find the current set through some other method.

Send transactions
~~~~~~~~~~~~~~~~~

Now we are ready to send some transactions. First Let's check the
balance of the two accounts we setup earlier:

..  code:: shelldown[6]

::

    ME=$(basecli keys get cool | awk '{print $2}')
    YOU=$(basecli keys get friend | awk '{print $2}')
    basecli query account $ME
    basecli query account $YOU

The first account is flush with cash, while the second account doesn't
exist. Let's send funds from the first account to the second:

..  code:: shelldown[7]

::

    basecli tx send --name=cool --amount=1000strings --to=$YOU --sequence=1

Now if we check the second account, it should have ``1000`` 'strings'
coins!

..  code:: shelldown[8]

::

    basecli query account $YOU

We can send some of these coins back like so:

..  code:: shelldown[9]

::

    basecli tx send --name=friend --amount=500strings --to=$ME --sequence=1

Note how we use the ``--name`` flag to select a different account to
send from.

If we try to send too much, we'll get an error:

..  code:: shelldown[10]

::

    basecli tx send --name=friend --amount=500000strings --to=$ME --sequence=2

Let's send another transaction:

..  code:: shelldown[11]

::

   basecli tx send --name=cool --amount=2345strings --to=$YOU --sequence=2

Note the ``hash`` value in the response - this is the hash of the
transaction. We can query for the transaction by this hash:

..  code:: shelldown[12]

::

    basecli query tx <HASH>

See ``basecli tx send --help`` for additional details.

Proof
-----

Even if you don't see it in the UI, the result of every query comes with
a proof. This is a Merkle proof that the result of the query is actually
contained in the state. And the state's Merkle root is contained in a
recent block header. Behind the scenes, ``countercli`` will not only
verify that this state matches the header, but also that the header is
properly signed by the known validator set. It will even update the
validator set as needed, so long as there have not been major changes
and it is secure to do so. So, if you wonder why the query may take a
second... there is a lot of work going on in the background to make sure
even a lying full node can't trick your client.

Accounts and Transactions
-------------------------

For a better understanding of how to further use the tools, it helps to
understand the underlying data structures.

Accounts
~~~~~~~~

The Basecoin state consists entirely of a set of accounts. Each account
contains a public key, a balance in many different coin denominations,
and a strictly increasing sequence number for replay protection. This
type of account was directly inspired by accounts in Ethereum, and is
unlike Bitcoin's use of Unspent Transaction Outputs (UTXOs). Note
Basecoin is a multi-asset cryptocurrency, so each account can have many
different kinds of tokens.

..  code:: golang

::

    type Account struct {
        PubKey   crypto.PubKey `json:"pub_key"` // May be nil, if not known.
        Sequence int           `json:"sequence"`
        Balance  Coins         `json:"coins"`
    }

    type Coins []Coin

    type Coin struct {
        Denom  string `json:"denom"`
        Amount int64  `json:"amount"`
    }

If you want to add more coins to a blockchain, you can do so manually in
the ``~/.basecoin/genesis.json`` before you start the blockchain for the
first time.

Accounts are serialized and stored in a Merkle tree under the key
``base/a/<address>``, where ``<address>`` is the address of the account.
Typically, the address of the account is the 20-byte ``RIPEMD160`` hash
of the public key, but other formats are acceptable as well, as defined
in the `Tendermint crypto
library <https://github.com/tendermint/go-crypto>`__. The Merkle tree
used in Basecoin is a balanced, binary search tree, which we call an
`IAVL tree <https://github.com/tendermint/go-merkle>`__.

Transactions
~~~~~~~~~~~~

Basecoin defines a transaction type, the ``SendTx``, which allows tokens
to be sent to other accounts. The ``SendTx`` takes a list of inputs and
a list of outputs, and transfers all the tokens listed in the inputs
from their corresponding accounts to the accounts listed in the output.
The ``SendTx`` is structured as follows:

..  code:: golang

::

    type SendTx struct {
      Gas     int64      `json:"gas"`
      Fee     Coin       `json:"fee"`
      Inputs  []TxInput  `json:"inputs"`
      Outputs []TxOutput `json:"outputs"`
    }

    type TxInput struct {
      Address   []byte           `json:"address"`   // Hash of the PubKey
      Coins     Coins            `json:"coins"`     //
      Sequence  int              `json:"sequence"`  // Must be 1 greater than the last committed TxInput
      Signature crypto.Signature `json:"signature"` // Depends on the PubKey type and the whole Tx
      PubKey    crypto.PubKey    `json:"pub_key"`   // Is present iff Sequence == 0
    }

    type TxOutput struct {
      Address []byte `json:"address"` // Hash of the PubKey
      Coins   Coins  `json:"coins"`   //
    }

Note the ``SendTx`` includes a field for ``Gas`` and ``Fee``. The
``Gas`` limits the total amount of computation that can be done by the
transaction, while the ``Fee`` refers to the total amount paid in fees.
This is slightly different from Ethereum's concept of ``Gas`` and
``GasPrice``, where ``Fee = Gas x GasPrice``. In Basecoin, the ``Gas``
and ``Fee`` are independent, and the ``GasPrice`` is implicit.

In Basecoin, the ``Fee`` is meant to be used by the validators to inform
the ordering of transactions, like in Bitcoin. And the ``Gas`` is meant
to be used by the application plugin to control its execution. There is
currently no means to pass ``Fee`` information to the Tendermint
validators, but it will come soon...

Note also that the ``PubKey`` only needs to be sent for
``Sequence == 0``. After that, it is stored under the account in the
Merkle tree and subsequent transactions can exclude it, using only the
``Address`` to refer to the sender. Ethereum does not require public
keys to be sent in transactions as it uses a different elliptic curve
scheme which enables the public key to be derived from the signature
itself.

Finally, note that the use of multiple inputs and multiple outputs
allows us to send many different types of tokens between many different
accounts at once in an atomic transaction. Thus, the ``SendTx`` can
serve as a basic unit of decentralized exchange. When using multiple
inputs and outputs, you must make sure that the sum of coins of the
inputs equals the sum of coins of the outputs (no creating money), and
that all accounts that provide inputs have signed the transaction.

Clean Up
--------

**WARNING:** Running these commands will wipe out any existing
information in both the ``~/.basecli`` and ``~/.basecoin`` directories,
including private keys.

To remove all the files created and refresh your environment (e.g., if
starting this tutorial again or trying something new), the following
commands are run:

..  code:: shelldown[end-of-tutorials]

::

    basecli reset_all
    rm -rf ~/.basecoin

In this guide, we introduced the ``basecoin`` and ``basecli`` tools,
demonstrated how to start a new basecoin blockchain and how to send
tokens between accounts, and discussed the underlying data types for
accounts and transactions, specifically the ``Account`` and the
``SendTx``.
