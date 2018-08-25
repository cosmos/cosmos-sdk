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

* Gaia
    * Make the transient store key use a distinct store key. [#2013](https://github.com/cosmos/cosmos-sdk/pull/2013)
    * [x/stake] \#1901 Validator type's Owner field renamed to Operator; Validator's GetOwner() renamed accordingly to comply with the SDK's Validator interface.
    * [docs] [#2001](https://github.com/cosmos/cosmos-sdk/pull/2001) Update slashing spec for slashing period
    * [x/stake, x/slashing] [#1305](https://github.com/cosmos/cosmos-sdk/issues/1305) - Rename "revoked" to "jailed"
    
* SDK
    * [core] \#1807 Switch from use of rational to decimal
    * [types] \#1901 Validator interface's GetOwner() renamed to GetOperator()
    * [types] \#2119 Parsed error messages and ABCI log errors to make them more human readable.

* Tendermint


FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [lcd] Endpoints to query staking pool and params

* Gaia CLI  (`gaiacli`)
  * [cli] Cmds to query staking pool and params
  * [gov][cli] #2062 added `--proposal` flag to `submit-proposal` that allows a JSON file containing a proposal to be passed in

* Gaia

* SDK
  * [querier] added custom querier functionality, so ABCI query requests can be handled by keepers

* Tendermint


IMPROVEMENTS

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

* Tendermint


BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
    * [cli] \#1997 Handle panics gracefully when `gaiacli stake {delegation,unbond}` fail to unmarshal delegation.

* Gaia

* SDK
    * \#1988 Make us compile on OpenBSD (disable ledger) [#1988] (https://github.com/cosmos/cosmos-sdk/issues/1988)
    * \#2105 Fix DB Iterator leak, which may leak a go routine.

* Tendermint
