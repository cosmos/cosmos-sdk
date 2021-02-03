<!--
order: 2
-->

# State

## FeeAllowance

Fee Allowances are identified by combining `Grantee` (The account address of fee allowance grantee) with the `Granter` (The account address of fee allowance granter).

Fee allowances are stored in the state as follows:

- FeeAllowance: `0x00 | grantee | granter -> ProtocolBuffer(FeeAllowance)`

+++ https://github.com/cosmos/cosmos-sdk/blob/master/x/feegrant/types/feegrant.pb.go#L358-L363