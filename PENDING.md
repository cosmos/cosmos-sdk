## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [lcd] https://github.com/cosmos/cosmos-sdk/pull/3045 Fix quoted json return on GET /keys (keys list)
  * [gaia-lite] [\#2191](https://github.com/cosmos/cosmos-sdk/issues/2191) Split `POST /stake/delegators/{delegatorAddr}/delegations` into `POST /stake/delegators/{delegatorAddr}/delegations`, `POST /stake/delegators/{delegatorAddr}/unbonding_delegations` and `POST /stake/delegators/{delegatorAddr}/redelegations`
  * [gaia-lite] [\#3056](https://github.com/cosmos/cosmos-sdk/pull/3056) `generate_only` and `simulate` have moved from query arguments to POST requests body.

* Gaia CLI  (`gaiacli`)

* Gaia

* SDK

* Tendermint


FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
  * \#2399 Implement `/slashing/parameters` endpoint to query slashing parameters.

* Gaia CLI  (`gaiacli`)
  * \#2399 Implement `params` command to query slashing parameters.

* Gaia

* SDK
  - [\#2926](https://github.com/cosmos/cosmos-sdk/issues/2926) Add TxEncoder to client TxBuilder.

* Tendermint


IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia
  * [\#3021](https://github.com/cosmos/cosmos-sdk/pull/3021) Add `--gentx-dir` to `gaiad collect-gentxs` to specify a directory from which collect and load gentxs.
    Add `--output-document` to `gaiad init` to allow one to redirect output to file.

* SDK

* Tendermint


BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia

* SDK

* Tendermint
