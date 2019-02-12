## PENDING

BREAKING CHANGES

* Gaia REST API

* Gaia CLI

* Gaia

* SDK
 * \#3621 staking.GenesisState.Bonds -> Delegations
 * \#3592 Drop deprecated keybase implementation's
   New constructor in favor of a new
   crypto/keys.New(string, string) implementation that
   returns a lazy keybase instance. Remove client.MockKeyBase,
   superseded by crypto/keys.NewInMemory()

* Tendermint

FEATURES

* Gaia REST API

* Gaia CLI

* Gaia

* SDK

* Tendermint


IMPROVEMENTS

* Gaia REST API

* Gaia CLI

* Gaia

* SDK
  * [\#3604] Improve SDK funds related error messages and allow for unicode in
  JSON ABCI log.
  * \#3621 remove many inter-module dependancies

* Tendermint


BUG FIXES

* Gaia REST API

* Gaia CLI
  * [\#3586](https://github.com/cosmos/cosmos-sdk/pull/3586) Incomplete ledger derivation paths in keybase 

* Gaia
  * [\#3585] Fix setting the tx hash in `NewResponseFormatBroadcastTxCommit`.

* SDK
  * [\#3582](https://github.com/cosmos/cosmos-sdk/pull/3582) Running `make test_unit was failing due to a missing tag

* Tendermint
