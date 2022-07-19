<!--
order: 5
-->

# Parameters

The bank module contains the following parameters:

| Key                | Type          | Example      |
| ------------------ | ------------- |--------------|
| SendEnabled        | []SendEnabled | (deprecated) |
| DefaultSendEnabled | bool          | true         |

## SendEnabled

The SendEnabled parameter is now deprecated and not to be use. It is replaced
with state store records.


## DefaultSendEnabled

The default send enabled value controls send transfer capability for all
coin denominations unless specifically included in the array of `SendEnabled`
parameters.
