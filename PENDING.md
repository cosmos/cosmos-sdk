## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
    * [x/stake] Validator.Owner renamed to Validator.Operator

* Gaia CLI  (`gaiacli`)
    * [x/stake] Validator.Owner renamed to Validator.Operator
    * [cli] unsafe_reset_all, show_validator, and show_node_id have been renamed to unsafe-reset-all, show-validator, and show-node-id
    * [cli] \#1983 --print-response now defaults to true in commands that create and send a transaction
    * [cli] \#1983 you can now pass --pubkey or --address to gaiacli keys show to return a plaintext representation of the key's address or public key for use with other commands
    * [cli] \#2061 changed proposalID in governance REST endpoints to proposal-id
    * [cli] \#2014 `gaiacli advanced` no longer exists - to access `ibc`, `rest-server`, and `validator-set` commands use `gaiacli ibc`, `gaiacli rest-server`, and `gaiacli tendermint`, respectively
    * [makefile] `get_vendor_deps` no longer updates lock file it just updates vendor directory. Use `update_vendor_deps` to update the lock file. [#2152](https://github.com/cosmos/cosmos-sdk/pull/2152)
    * [cli] \#2190 `gaiacli init --gen-txs` is now `gaiacli init --with-txs` to reduce confusion

* Gaia
    * Make the transient store key use a distinct store key. [#2013](https://github.com/cosmos/cosmos-sdk/pull/2013)
    * [x/stake] \#1901 Validator type's Owner field renamed to Operator; Validator's GetOwner() renamed accordingly to comply with the SDK's Validator interface.
    * [docs] [#2001](https://github.com/cosmos/cosmos-sdk/pull/2001) Update slashing spec for slashing period
    * [x/stake, x/slashing] [#1305](https://github.com/cosmos/cosmos-sdk/issues/1305) - Rename "revoked" to "jailed"
    
* SDK
    * [core] \#1807 Switch from use of rational to decimal
    * [types] \#1901 Validator interface's GetOwner() renamed to GetOperator()
    * [types] \#2119 Parsed error messages and ABCI log errors to make them more human readable.
    * [simulation] Rename TestAndRunTx to Operation [#2153](https://github.com/cosmos/cosmos-sdk/pull/2153)

* Tendermint


FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [lcd] Endpoints to query staking pool and params

* Gaia CLI  (`gaiacli`)
  * [cli] Cmds to query staking pool and params
  * [gov][cli] #2062 added `--proposal` flag to `submit-proposal` that allows a JSON file containing a proposal to be passed in
  * [cli] \#2047 Setting the --gas flag value to 0 triggers a simulation of the tx before the actual execution. The gas estimate obtained via the simulation will be used as gas limit in the actual execution.
  * [cli] \#2047 The --gas-adjustment flag can be used to adjust the estimate obtained via the simulation triggered by --gas=0.

* Gaia

* SDK
  * [querier] added custom querier functionality, so ABCI query requests can be handled by keepers
  * [simulation] \#1924 allow operations to specify future operations

* Tendermint


IMPROVEMENTS
* [tools] Improved terraform and ansible scripts for infrastructure deployment
* [tools] Added ansible script to enable process core dumps

* Gaia REST API (`gaiacli advanced rest-server`)
    * [x/stake] \#2000 Added tests for new staking endpoints

* Gaia CLI  (`gaiacli`)
    * [cli] #2060 removed `--select` from `block` command
    * [cli] #2128 fixed segfault when exporting directly after `gaiad init`

* Gaia
    * [x/stake] [#2023](https://github.com/cosmos/cosmos-sdk/pull/2023) Terminate iteration loop in `UpdateBondedValidators` and `UpdateBondedValidatorsFull` when the first revoked validator is encountered and perform a sanity check.
    * [x/auth] Signature verification's gas cost now accounts for pubkey type. [#2046](https://github.com/tendermint/tendermint/pull/2046)

* SDK
    * [tools] Make get_vendor_deps deletes `.vendor-new` directories, in case scratch files are present.
    * [cli] \#1632 Add integration tests to ensure `basecoind init && basecoind` start sequences run successfully for both `democoin` and `basecoin` examples.
    * [store] Speedup IAVL iteration, and consequently everything that requires IAVL iteration. [#2143](https://github.com/cosmos/cosmos-sdk/issues/2143)
    * [simulation] Make timestamps randomized [#2153](https://github.com/cosmos/cosmos-sdk/pull/2153)

* Tendermint

BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
    * [cli] \#1997 Handle panics gracefully when `gaiacli stake {delegation,unbond}` fail to unmarshal delegation.

* Gaia

* SDK
    * \#1988 Make us compile on OpenBSD (disable ledger) [#1988] (https://github.com/cosmos/cosmos-sdk/issues/1988)
    * \#2105 Fix DB Iterator leak, which may leak a go routine.
    * [ledger] \#2064 Fix inability to sign and send transactions via the LCD by
    loading a Ledger device at runtime.
    * \#2158 Fix non-deterministic ordering of validator iteration when slashing in `gov EndBlocker`

* Tendermint
