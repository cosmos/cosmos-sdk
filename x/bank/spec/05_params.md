<!--
order: 5
-->

# Parameters

The bank module contains the following parameters:

| Key                | Type          | Example                            |
| ------------------ | ------------- | ---------------------------------- |
| SendEnabled        | []SendEnabled | [{denom: "stake", enabled: true }] |
| DefaultSendEnabled | bool          | true                               |

## SendEnabled

The send enabled parameter is an array of SendEnabled entries mapping coin
denominations to their send_enabled status. Entries in this list take
precedence over the `DefaultSendEnabled` setting.

## DefaultSendEnabled

The default send enabled value controls send transfer capability for all
coin denominations unless specifically included in the array of `SendEnabled`
parameters.
