<!--
order: 2
-->

# State

The `x/quarantine` module uses key/value pairs to store quarantine-related data in state.

## Quarantined Accounts

When an account opts into quarantine, the following record is made:

```
0x00 | len([]byte(<account address>)) | []byte(<account address>) -> 0x00
```

When an account opts out of quarantine, that record is deleted.

## Auto-Responses

Auto-Responses are stored using the following format:

```
0x01 | len([]byte(<receiver address>)) | []byte(<receiver address>) | len([]byte(<sender address>)) | []byte(<sender address>) -> <response> 
```

`<response>` values:
- `0x01` = `AUTO_RESPONSE_ACCEPT`
- `0x02` = `AUTO_RESPONSE_DECLINE`

Instead of storing `AUTO_RESPONSE_UNSPECIFIED` the record is deleted.

## Quarantine Records

Records of quarantined funds are stored using the following format:

```
0x02 | len([]byte(<receiver address>)) | []byte(<receiver address>) | len([]byte(<record suffix>)) | []byte(<record suffix>) -> ProtocolBuffer(QuarantineRecord) 
```

When there is a single sender, the `<record suffix>` is the `<sender address>`.

When there are multiple senders, the `<record suffix>` is a function of all sender addresses combined.
Specifically, all involved sender addresses are sorted and concatenated into a single `[]byte`, then provided to a `sha256` checksum generator.

Once quarantined funds are accepted and released, this record is deleted.

## Quarantine Records Suffix Index

When there are multiple senders, an index entry is made for each sender.
These entries use the following format:

```
0x03 | len([]byte(<receiver address>)) | []byte(<receiver address>) | len([]byte(<sender address>)) | []byte(<sender address>) -> ProtocolBuffer(QuarantineRecordSuffixIndex)
```

These entries allow multi-sender quarantine records to be located based on a single sender.
They are not needed for single-sender records; as such, they are only made for multi-sender records. 

Once a quarantine record is deleted, its suffix index entries are also deleted.
