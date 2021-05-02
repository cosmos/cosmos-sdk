<!--
order: 2
-->

# State

## FeeAllowance

Fee Allowances are identified by combining `Grantee` (the account address of fee allowance grantee) with the `Granter` (the account address of fee allowance granter).

Fee allowances are stored in the state as follows:

- FeeAllowance: `0x00 | grantee_addr_len (1 byte) | grantee_addr_bytes |  granter_addr_len (1 byte) | granter_addr_bytes -> ProtocolBuffer(FeeAllowance)`

+++ https://github.com/cosmos/cosmos-sdk/blob/d97e7907f176777ed8a464006d360bb3e1a223e4/x/feegrant/types/feegrant.pb.go#L358-L363
