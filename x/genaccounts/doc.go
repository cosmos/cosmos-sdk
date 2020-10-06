/*
Package genaccounts is now deprecated.

IMPORTANT: This module has been replaced by ADR 011: Generalize Module Accounts.
The ADR can be found here: https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-011-generalize-genesis-accounts.md.

Genesis accounts that existed in the genesis application state under `app_state.accounts`
now exists under the x/auth module's genesis state under the `app_state.auth.accounts` key.
Migration can be performed via x/auth/legacy/v038/migrate.go. In addition, because genesis
accounts are now generalized via an interface, it is now up to the application to
define the concrete types and the respective client logic to add them to a genesis
state/file. For an example implementation of the `add-genesis-account` command please
refer to https://github.com/cosmos/gaia/pull/122.
*/
package genaccounts
