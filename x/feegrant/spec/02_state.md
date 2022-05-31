<!--
order: 2
-->

# State

## FeeAllowance

Fee Allowances are identified by combining `Grantee` (the account address of fee allowance grantee) with the `Granter` (the account address of fee allowance granter).

Fee allowance grants are stored in the state as follows:

* Grant: `0x00 | grantee_addr_len (1 byte) | grantee_addr_bytes |  granter_addr_len (1 byte) | granter_addr_bytes -> ProtocolBuffer(Grant)`

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/feegrant/feegrant.pb.go#L221-L229

## FeeAllowanceQueue

Fee Allowances queue items are identified by combining the `FeeAllowancePrefixQueue` (i.e., 0x01), `expiration`, `grantee` (the account address of fee allowance grantee), `granter` (the account address of fee allowance granter). Endblocker checks `FeeAllowanceQueue` state for the expired grants and prunes them from  `FeeAllowance` if there are any found.

Fee allowance queue keys are stored in the state as follows:

* Grant: `0x01 | expiration_bytes | grantee_addr_len (1 byte) | grantee_addr_bytes |  granter_addr_len (1 byte) | granter_addr_bytes -> EmptyBytes`
