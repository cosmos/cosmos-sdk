Using Gaia
==========

The purpose of the ``gaia`` staking module is to provide users with the ability to 1) declare candidacy as a validator, 2) bond/unbond to a candidate.

For the time being, the ``gaia`` tooling is installed seperately from the Cosmos-SDK:

::

    go get github.com/cosmos/gaia
    cd $GOPATH/src/github.com/cosmos/gaia
    make get_vendor_deps
    make install

The ``gaia`` tool has three primary commands:

::

    Available Commands:
      node        The Cosmos Network delegation-game blockchain test
      rest-server REST client for gaia commands
      client      Gaia light client
                        
      version     Show version info
      help        Help about any command

and a handful of flags that are highlighted only as necessary.

The ``gaia node`` command is a proxt for running a tendermint node. You'll be using
this command to either initialize a new node, or - using existing files - joining
the testnet. 

The ``gaia rest-server`` command is used by the `cosmos UI <https://github.com/cosmos/cosmos-ui>`__.

Lastly, the ``gaia client`` command is the workhorse of the staking module. It allows
for sending various transactions and other types of interaction with a running chain.
