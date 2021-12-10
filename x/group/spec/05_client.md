<!--
order: 6
-->

# Client

## CLI

A user can query and interact with the `group` module using the CLI.

### Query

The `query` commands allow users to query `group` state.

```bash
simd query group --help
```

#### group-info

The `group-info` command allows users to query for group info by given group id.

```bash
simd query group group-info [id] [flags]
```

Example:

```bash
simd query group group-info 1
```

Example Output:

```bash
admin: cosmos1..
group_id: "1"
metadata: null
total_weight: "3"
version: "1"
```

#### group-accounts-by-admin

The `group-accounts-by-admin` command allows users to query for group accounts by admin account address with pagination flags.

```bash
simd query group group-accounts-by-admin [admin] [flags]
```

Example:

```bash
simd query group group-accounts-by-admin cosmos1..
```

Example Output:

```bash
group_accounts:
- address: cosmos1..
  admin: cosmos1..
  decision_policy:
    '@type': /regen.group.v1alpha1.ThresholdDecisionPolicy
    threshold: "1"
    timeout: 1s
  derivation_key: AgAAAAAAAAA=
  group_id: "1"
  metadata: metadata
  version: "1"
- address: cosmos1..
  admin: cosmos1..
  decision_policy:
    '@type': /regen.group.v1alpha1.ThresholdDecisionPolicy
    threshold: "1"
    timeout: 1s
  derivation_key: AQAAAAAAAAA=
  group_id: "1"
  metadata: null
  version: "1"
pagination:
  next_key: null
  total: "2"
```

#### group-account-info

The `group-account-info` command allows users to query  for group account info by group account address

```bash
simd query group group-account-info [group-account] [flags]
```

Example:

```bash
simd query group group-account-info cosmos1..
```

Example Output:

```bash
address: cosmos1..
admin: cosmos1..
decision_policy:
  '@type': /regen.group.v1alpha1.ThresholdDecisionPolicy
  threshold: "1"
  timeout: 1s
derivation_key: AgAAAAAAAAA=
group_id: "1"
metadata: metadata
version: "1"
```

#### group-accounts-by-group

The `group-accounts-by-group` command allows users to query for group accounts by group id with pagination flags.

```bash
simd query group group-accounts-by-group [group-id] [flags]
```

Example:

```bash
simd query group group-accounts-by-group 1 
```

Example Output:

```bash
group_accounts:
- address: cosmos1..
  admin: cosmos1..
  decision_policy:
    '@type': /regen.group.v1alpha1.ThresholdDecisionPolicy
    threshold: "1"
    timeout: 1s
  derivation_key: AgAAAAAAAAA=
  group_id: "1"
  metadata: metadata
  version: "1"
- address: cosmos1..
  admin: cosmos1..
  decision_policy:
    '@type': /regen.group.v1alpha1.ThresholdDecisionPolicy
    threshold: "1"
    timeout: 1s
  derivation_key: AQAAAAAAAAA=
  group_id: "1"
  metadata: null
  version: "1"
pagination:
  next_key: null
  total: "2"
```

#### groups-by-admin

The `groups-by-admin` command allows users to query for groups by admin account address with pagination flags.

```bash
simd query group groups-by-admin [admin] [flags]
```

Example:

```bash
simd query group groups-by-admin cosmos1..
```

Example Output:

```bash
groups:
- admin: regen15ehk66keu65g3vnce0k0w2a7c72gzmlufn39aa
  group_id: "1"
  metadata: null
  total_weight: "3"
  version: "1"
- admin: regen15ehk66keu65g3vnce0k0w2a7c72gzmlufn39aa
  group_id: "2"
  metadata: null
  total_weight: "3"
  version: "1"
pagination:
  next_key: null
  total: "2"
```

#### group-members

The `group-members` command allows users to query for group members by group id with pagination flags.

```bash
simd query group group-members [id] [flags]
```

Example:

```bash
simd query group group-members 1
```

Example Output:

```bash
members:
- group_id: "1"
  member:
    address: cosmos1..
    metadata: null
    weight: "2"
- group_id: "1"
  member:
    address: cosmos1..
    metadata: null
    weight: "1"
pagination:
  next_key: null
  total: "2"
```








### Transactions

The `tx` commands allow users to interact with the `group` module.

```bash
simd tx group --help
```

#### create-group

The `create-group` command allows users to create a group which is an aggregation of member accounts with associated weights and
an administrator account.

```bash
simd tx group create-group [admin] [metadata] [members-json-file]
```

Example:

```bash
simd tx group create-group cosmos1.. metadata members.json 
```

#### create-group-account

The `create-group-account` command allows users to create a group account which is an account associated with a group and a decision policy. 

```bash
simd tx group create-group-account [admin] [group-id] [metadata] [decision-policy] [flags]
```

Example:

```bash
simd tx group create-group-account cosmos1.. 1 metadata '{"@type":"/regen.group.v1alpha1.ThresholdDecisionPolicy", "threshold":"1", "timeout":"1s"}' 
```

#### update-group-admin

The `update-group-admin` command allows users to update a group's admin.

```bash
simd tx group update-group-admin [admin] [group-id] [new-admin] [flags]
```

Example:

```bash
simd tx group update-group-admin cosmos1.. 1 cosmos1..
```

#### update-group-members

The `update-group-members` command allows users to update a group's members.

```bash
simd tx group update-group-members [admin] [group-id] [members-json-file] [flags]
```

Example:

```bash
simd tx group update-group-members cosmos1.. 1 members.json
```

#### update-group-metadata

The `update-group-metadata` command allows users to update a group's metadata.

```bash
simd tx group update-group-metadata [admin] [group-id] [metadata] [flags]
```

Example:

```bash
simd tx group update-group-metadata cosmos1.. 1 newmetadata
```

#### update-group-account-admin

The `update-group-account-admin` command allows users to update a group account admin.

```bash
simd tx group update-group-account-admin [admin] [group-account] [new-admin] [flags]
```

Example:

```bash
simd tx group update-group-account-admin cosmos1.. cosmos1.. cosmos1..
```

#### update-group-account-metadata

The `update-group-account-metadata` command allows users to update a group account metadata.

```bash
simd tx group update-group-account-metadata [admin] [group-account] [new-metadata] [flags]
```

Example:

```bash
simd tx group update-group-account-metadata cosmos1.. cosmos1.. newmetadata
```

#### update-group-account-policy

The `update-group-account-policy` command allows users to update a group account decision policy.

```bash
simd  tx group update-group-account-policy [admin] [group-account] [decision-policy] [flags]
```

Example:

```bash
simd tx group update-group-account-policy cosmos1.. cosmos1.. '{"@type":"/regen.group.v1alpha1.ThresholdDecisionPolicy", "threshold":"2", "timeout":"2s"}' 
```

#### create-proposal

The `create-proposal` command allows users to submit a new proposal.

```bash
simd tx group create-proposal [group-account] [proposer[,proposer]*] [msg_tx_json_file] [metadata] [flags]
```

Example:

```bash
simd tx group create-proposal cosmos1.. cosmos1.. msg_tx.json metadata
```

#### exec

The `exec` command allows users to execute a proposal.

```bash
simd tx group exec [proposal-id] [flags]
```

Example:

```bash
simd tx group exec 1
```
