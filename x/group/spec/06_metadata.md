<!--
order: 6
-->

# Metadata

The group module has four locations for metadata where users can provide further context about the on-chain actions they are taking. By default all metadata fields have a 255 character length field where metadata can be stored in json format, either on-chain or off-chain depending on the amount of data required. Here we provide a recommendation for the json structure and where the data should be stored. There are two important factors in making these recommendations. First, that the group and gov modules are consistent with one another, note the number of proposals made by all groups may be quite large. Second, that client applications such as block explorers and governance interfaces have confidence in the consistency of metadata structure accross chains.

## Proposal

Location: off-chain as json object stored on IPFS (mirrors [gov proposal](../../gov/spec/08_metadata.md#proposal))

```json
{
  "title": "",
  "authors": "",
  "summary": "",
  "details": "",
  "proposal_forum_url": "",
  "vote_option_context": "",
}
```

## Vote

Location: on-chain as json within 255 character limit (mirrors [gov vote](../../gov/spec/08_metadata.md#vote))

```json
{
  "justification": "",
}
```

## Group

Location: off-chain as json object stored on IPFS

```json
{
  "name": "",
  "description": "",
  "group_website_url": "",
  "group_forum_url": "",
}
```

## Decision policy

Location: on-chain as json within 255 character limit

```json
{
  "name": "",
  "description": "",
}
```
