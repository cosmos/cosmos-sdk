<!--
order: 5
-->

# Parameters

The bank module contains the following parameters:

| Key         | Type          | Example                            |
|-------------|---------------|------------------------------------|
| SendEnabled | []SendEnabled | [{denom: "stake", enabled: true }] |

## SendEnabled

The send enabled parameter is an array of SendEnabled entries mapping coin
denominations to their send_enabled status.  An empty string for denomination
indicates the default value to use for all unspecified denominations.