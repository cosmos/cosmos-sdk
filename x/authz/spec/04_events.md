<!--
order: 4
-->

# Events

The authz module emits the following events:

## Keeper

### GrantAuthorization

| Type                 | Attribute Key     | Attribute Value    |
|----------------------|-------------------|--------------------|
| grant-authorization  | module            | authz              |
| grant-authorization  | grant-type        | {msgType}          |
| grant-authorization  | granter           | {granterAddress}   |
| grant-authorization  | grantee           | {granteeAddress}   |


### RevokeAuthorization

| Type                 | Attribute Key     | Attribute Value    |
|----------------------|-------------------|--------------------|
| revoke-authorization | module            | authz              |
| revoke-authorization | grant-type        | {msgType}          |
| revoke-authorization | granter           | {granterAddress}   |
| revoke-authorization | grantee           | {granteeAddress}   |
