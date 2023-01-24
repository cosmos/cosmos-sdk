<!--
order: 2
-->

# State

The `x/sanction` module uses key/value pairs to store sanction-related data in state.

## Params

Each param field is stored in its own record with this format:

```
0x00 | []byte(<param name>) -> []byte(<param value>)
```

| Param Field                      | `<param name>`                     | `<param value>` format |
|----------------------------------|------------------------------------|------------------------|
| `ImmediateSanctionMinDeposit`    | `immediate_sanction_min_deposit`   | `sdk.Coins.String()`   |
| `ImmediateUnsanctionMinDeposit`  | `immediate_unsanction_min_deposit` | `sdk.Coins.String()`   |

## Sanctioned Accounts

When an account is sanctioned, the following record is made:

```
0x01 | len([]byte(<account address>)) | []byte(<account address>) -> 0x01
```

When an account is unsanctioned, that record is deleted.

## Temporary Entries

Immediate temporary sanctions and/or unsanctions are enacted by creating the following record:

```
0x02 | len([]byte(<account address>)) | []byte(<account address>) | [8]byte(<gov prop id>) -> byte(<value>)
```

| Entry type | `<value>` |
|------------|-----------|
| Sanction   | `0x01`    |
| Unsanction | `0x00`    |

When an account is sanctioned or unsanctioned, all temporary entry records for the address are removed.
If a proposal does not pass, all temporary entry records for that proposal are removed.

## Temporary Index

When a temporary entry is created, the following index record is also created:

```
0x03 | [8]byte(<gov prop id>) | len([]byte(<account address>)) | []byte(<account address>) -> byte(<value>)
```

The same `<value>` is used as the correlated temporary entry.

Temporary index records are removed when their correlated temporary entry record is removed.
