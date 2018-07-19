## PENDING

BREAKING CHANGES
* [baseapp] Msgs are no longer run on CheckTx, removed `ctx.IsCheckTx()`
* [x/stake] Fixed the period check for the inflation calculation
* [baseapp] NewBaseApp constructor now takes sdk.TxDecoder as argument instead of wire.Codec
* [x/auth] Default TxDecoder can be found in `x/auth` rather than baseapp

FEATURES
* [lcd] Can now query governance proposals by ProposalStatus

IMPROVEMENTS
* [baseapp] Allow any alphanumeric character in route
* [cli] Improve error messages for all txs when the account doesn't exist
* [tools] Remove `rm -rf vendor/` from `make get_vendor_deps`
* [x/stake] Add revoked to human-readable validator
* [x/auth] Recover ErrorOutOfGas panic in order to set sdk.Result attributes correctly

BUG FIXES
*  \#1666 Add intra-tx counter to the genesis validators
