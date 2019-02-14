## PENDING

BREAKING CHANGES

* Gaia REST API

* Gaia CLI

* Gaia

* SDK
 * \#3580 Migrate HTTP request/response types and utilities to types/rest.
 * \#3592 Drop deprecated keybase implementation's New() constructor in
   favor of a new crypto/keys.New(string, string) implementation that
   returns a lazy keybase instance. Remove client.MockKeyBase,
   superseded by crypto/keys.NewInMemory()
 * \#3621 staking.GenesisState.Bonds -> Delegations

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
  * \#3621 remove many inter-module dependancies
  * [\#3601] JSON-stringify the ABCI log response which includes the log and message
  index.
  * [\#3604] Improve SDK funds related error messages and allow for unicode in
  JSON ABCI log.
  * [\#3620](https://github.com/cosmos/cosmos-sdk/pull/3620) Version command shows build tags
  * [\#3638] Add Bcrypt benchmarks & justification of security parameter choice

* Tendermint
  * [\#3618] Upgrade to Tendermint 0.30.03

BUG FIXES

* Gaia REST API

* Gaia CLI

* Gaia

* SDK

* Tendermint
