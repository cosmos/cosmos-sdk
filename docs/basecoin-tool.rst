.. raw:: html

   <!--- shelldown script template, see github.com/rigelrozanski/shelldown
   #!/bin/bash

   testTutorial_BasecoinTool() {

       rm -rf ~/.basecoin
       rm -rf ~/.basecli
       rm -rf example-data
       KEYPASS=qwertyuiop

       (echo $KEYPASS; echo $KEYPASS) | #shelldown[0][0] >/dev/null ; assertTrue "Expected true for line $LINENO" $? 
       #shelldown[0][1] >/dev/null ; assertTrue "Expected true for line $LINENO" $?
       #shelldown[1][0] ; assertTrue "Expected true for line $LINENO" $? 
       #shelldown[1][1] ; assertTrue "Expected true for line $LINENO" $? 
       
       #shelldown[1][2] >>/dev/null 2>&1 &
       sleep 5 ; PID_SERVER=$! ; disown ; assertTrue "Expected true for line $LINENO" $?
       kill -9 $PID_SERVER >/dev/null 2>&1 ; sleep 1
       
       #shelldown[2][0] ; assertTrue "Expected true for line $LINENO" $? 
       #shelldown[2][1] >>/dev/null 2>&1 &
       sleep 5 ; PID_SERVER=$! ; disown ; assertTrue "Expected true for line $LINENO" $?
       kill -9 $PID_SERVER >/dev/null 2>&1 ; sleep 1
       
       #shelldown[3][-1] >/dev/null ; assertTrue "Expected true for line $LINENO" $? 
       
       #shelldown[4][-1] >>/dev/null 2>&1 &
       sleep 5 ; PID_SERVER=$! ; disown ; assertTrue "Expected true for line $LINENO" $?
       #shelldown[5][-1] >>/dev/null 2>&1 &
       sleep 5 ; PID_SERVER2=$! ; disown ; assertTrue "Expected true for line $LINENO" $?
       kill -9 $PID_SERVER $PID_SERVER2 >/dev/null 2>&1 ; sleep 1
       
       #shelldown[4][-1] >>/dev/null 2>&1 &
       sleep 5 ; PID_SERVER=$! ; disown ; assertTrue "Expected true for line $LINENO" $?
       #shelldown[6][0] ; assertTrue "Expected true for line $LINENO" $? 
       #shelldown[6][1] >>/dev/null 2>&1 &
       sleep 5 ; PID_SERVER2=$! ; disown ; assertTrue "Expected true for line $LINENO" $?
       kill -9 $PID_SERVER $PID_SERVER2 >/dev/null 2>&1 ; sleep 1
       
       #shelldown[7][-1] >/dev/null ; assertTrue "Expected true for line $LINENO" $? 
       #shelldown[8][-1] >/dev/null ; assertTrue "Expected true for line $LINENO" $?
       (echo $KEYPASS; echo $KEYPASS) | #shelldown[9][-1] >/dev/null ; assertTrue "Expected true for line $LINENO" $? 
       #shelldown[10][-1] >/dev/null ; assertTrue "Expected true for line $LINENO" $? 
       #shelldown[11][-1] >/dev/null ; assertTrue "Expected true for line $LINENO" $? 
      
       #cleanup 
       rm -rf example-data
   }

   # load and run these tests with shunit2!
   DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
   . $DIR/shunit2
   -->

Basecoin The Tool
=================

We previously learned about basecoin basics. In this tutorial, we
provide more details on using the Basecoin tool.

Generate a Key
--------------

Generate a key using the ``basecli`` tool:

.. comment code:: shelldown[0]

::

    basecli keys new mykey
    ME=$(basecli keys get mykey | awk '{print $2}')

Data Directory
--------------

By default, ``basecoin`` works out of ``~/.basecoin``. To change this,
set the ``BCHOME`` environment variable:

.. comment code:: shelldown[1]

::

    export BCHOME=~/.my_basecoin_data
    basecoin init $ME
    basecoin start

or

.. comment code:: shelldown[2]

::

    BCHOME=~/.my_basecoin_data basecoin init $ME
    BCHOME=~/.my_basecoin_data basecoin start

ABCI Server
-----------

So far we have run Basecoin and Tendermint in a single process. However,
since we use ABCI, we can actually run them in different processes.
First, initialize them:

.. comment code:: shelldown[3]

::

    basecoin init $ME

This will create a single ``genesis.json`` file in ``~/.basecoin`` with
the information for both Basecoin and Tendermint.

Now, In one window, run

.. comment code:: shelldown[4]

::

    basecoin start --without-tendermint

and in another,

.. comment code:: shelldown[5]

::

    TMROOT=~/.basecoin tendermint node

You should see Tendermint start making blocks!

Alternatively, you could ignore the Tendermint details in
``~/.basecoin/genesis.json`` and use a separate directory by running:

.. comment code:: shelldown[6]

::

    tendermint init
    tendermint node

See the `tendermint documentation <https://tendermint.readthedocs.io>`__ for more information.

Keys and Genesis
----------------

In previous tutorials we used ``basecoin init`` to initialize
``~/.basecoin`` with the default configuration. This command creates
files both for Tendermint and for Basecoin, and a single
``genesis.json`` file for both of them. You can read more about these
files in the Tendermint documentation.

Now let's make our own custom Basecoin data.

First, create a new directory:

.. comment code:: shelldown[7]

::

    mkdir example-data

We can tell ``basecoin`` to use this directory by exporting the
``BCHOME`` environment variable:

.. comment code:: shelldown[8]

::

    export BCHOME=$(pwd)/example-data

If you're going to be using multiple terminal windows, make sure to add
this variable to your shell startup scripts (eg. ``~/.bashrc``).

Now, let's create a new key:

.. comment code:: shelldown[9]

::

    basecli keys new foobar

The key's info can be retrieved with

.. comment code:: shelldown[10]

::

    basecli keys get foobar -o=json

You should get output which looks similar to the following:

.. code:: json

    {
      "name": "foobar",
      "address": "404C5003A703C7DA888C96A2E901FCE65A6869D9",
      "pubkey": {
        "type": "ed25519",
        "data": "8786B7812AB3B27892D8E14505EEFDBB609699E936F6A4871B1983F210736EEA"
      }
    }

Yours will look different - each key is randomly derived. Now we can
make a ``genesis.json`` file and add an account with our public key:

.. code:: json

    {
      "app_hash": "",
      "chain_id": "example-chain",
      "genesis_time": "0001-01-01T00:00:00.000Z",
      "validators": [
        {
          "amount": 10,
          "name": "",
          "pub_key": {
            "type": "ed25519",
            "data": "7B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
          }
        }
      ],
      "app_options": {
        "accounts": [
          {
            "pub_key": {
              "type": "ed25519",
              "data": "8786B7812AB3B27892D8E14505EEFDBB609699E936F6A4871B1983F210736EEA"
            },
            "coins": [
              {
                "denom": "gold",
                "amount": 1000000000
              }
            ]
          }
        ]
      }
    }

Here we've granted ourselves ``1000000000`` units of the ``gold`` token.
Note that we've also set the ``chain-id`` to be ``example-chain``. All
transactions must therefore include the ``--chain-id example-chain`` in
order to make sure they are valid for this chain. Previously, we didn't
need this flag because we were using the default chain ID
("test\_chain\_id"). Now that we're using a custom chain, we need to
specify the chain explicitly on the command line.

Note we have also left out the details of the Tendermint genesis. See the
`Tendermint documentation <https://tendermint.readthedocs.io>`__ for more
information.

Reset
-----

You can reset all blockchain data by running:

.. (comment) code:: shelldown[11]

::

    basecoin unsafe_reset_all

Similarly, you can reset client data by running:

.. (comment) code:: shelldown[12]

::

    basecli reset_all

Genesis
-------

Any required plugin initialization should be constructed using
``SetOption`` on genesis. When starting a new chain for the first time,
``SetOption`` will be called for each item the genesis file. Within
genesis.json file entries are made in the format:
``"<plugin>/<key>", "<value>"``, where ``<plugin>`` is the plugin name,
and ``<key>`` and ``<value>`` are the mycoin passed into the plugin
SetOption function. This function is intended to be used to set plugin
specific information such as the plugin state.
