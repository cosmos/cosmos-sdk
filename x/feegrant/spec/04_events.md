<!--
order: 4
-->

# Events

The feegrant module emits the following events:

# Message Servers

### MsgGrantFeeAllowance

| Type     | Attribute Key | Attribute Value    |
| -------- | ------------- | ------------------ |
| message  | action        | set_feegrant       |
| message  | granter       | {granterAddress}   |
| message  | grantee       | {granteeAddress}   |

### MsgRevokeFeeAllowance

| Type     | Attribute Key | Attribute Value    |
| -------- | ------------- | ------------------ |
| message  | action        | revoke_feegrant    |
| message  | granter       | {granterAddress}   |
| message  | grantee       | {granteeAddress}   |

### Exec fee allowance

| Type     | Attribute Key | Attribute Value    |
| -------- | ------------- | ------------------ |
| message  | action        | use_feegrant       |
| message  | granter       | {granterAddress}   |
| message  | grantee       | {granteeAddress}   |