<!--
order: 4
-->

# Events

The `x/sanction` module emits the following events:

## EventAddressSanctioned

This event is emitted when an account is sanctioned.

`@Type`: `/cosmos.sanction.v1beta1.EventAddressSanctioned`

| Attribute Key | Attribute Value                       |
|---------------|---------------------------------------|
| address       | {bech32 string of sanctioned account} |

## EventAddressUnsanctioned

This event is emitted when an account is unsanctioned.

`@Type`: `/cosmos.sanction.v1beta1.EventAddressUnsanctioned`

| Attribute Key | Attribute Value                         |
|---------------|-----------------------------------------|
| address       | {bech32 string of unsanctioned account} |

## EventTempAddressSanctioned

This event is emitted when a temporary sanction is placed on an account.

`@Type`: `/cosmos.sanction.v1beta1.EventTempAddressSanctioned`

| Attribute Key | Attribute Value                       |
|---------------|---------------------------------------|
| address       | {bech32 string of sanctioned account} |

## EventTempAddressUnsanctioned

This event is emitted when a temporary unsanction is placed on an account.

`@Type`: `/cosmos.sanction.v1beta1.EventTempAddressUnsanctioned`

| Attribute Key | Attribute Value                         |
|---------------|-----------------------------------------|
| address       | {bech32 string of unsanctioned account} |

## EventParamsUpdated

This event is emitted when the `x/sanction` module's params are updated.

`@Type`: `/cosmos.sanction.v1beta1.EventParamsUpdated`

| Attribute Key | Attribute Value |
|---------------|-----------------|
| (none)        |                 |
