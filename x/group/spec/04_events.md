<!--
order: 4
-->

# Events

The group module emits the following events:

## EventCreateGroup

| Type                                  | Attribute Key | Attribute Value                       |
|---------------------------------------|---------------|---------------------------------------|
| message                               | action        | /cosmos.group.v1beta1.Msg/CreateGroup |
| cosmos.group.v1beta1.EventCreateGroup | group_id      | {groupId}                             |

## EventUpdateGroup

| Type                                  | Attribute Key | Attribute Value                                                 |
|---------------------------------------|---------------|-----------------------------------------------------------------|
| message                               | action        | /cosmos.group.v1beta1.Msg/UpdateGroup{Admin\|Metadata\|Members} |
| cosmos.group.v1beta1.EventUpdateGroup | group_id      | {groupId}                                                       |

## EventCreateGroupAccount

| Type                                         | Attribute Key | Attribute Value                              |
|----------------------------------------------|---------------|----------------------------------------------|
| message                                      | action        | /cosmos.group.v1beta1.Msg/CreateGroupAccount |
| cosmos.group.v1beta1.EventCreateGroupAccount | address       | {groupAccountAddress}                        |

## EventUpdateGroupAccount

| Type                                         | Attribute Key | Attribute Value                                                               |
|----------------------------------------------|---------------|-------------------------------------------------------------------------------|
| message                                      | action        | /cosmos.group.v1beta1.Msg/UpdateGroupAccount{Admin\|Metadata\|DecisionPolicy} |
| cosmos.group.v1beta1.EventUpdateGroupAccount | address       | {groupAccountAddress}                                                         |

## EventCreateProposal

| Type                                     | Attribute Key | Attribute Value                          |
|------------------------------------------|---------------|------------------------------------------|
| message                                  | action        | /cosmos.group.v1beta1.Msg/CreateProposal |
| cosmos.group.v1beta1.EventCreateProposal | proposal_id   | {proposalId}                             |

## EventVote

| Type                           | Attribute Key | Attribute Value                |
|--------------------------------|---------------|--------------------------------|
| message                        | action        | /cosmos.group.v1beta1.Msg/Vote |
| cosmos.group.v1beta1.EventVote | proposal_id   | {proposalId}                   |

## EventExec

| Type                           | Attribute Key | Attribute Value                |
|--------------------------------|---------------|--------------------------------|
| message                        | action        | /cosmos.group.v1beta1.Msg/Exec |
| cosmos.group.v1beta1.EventExec | proposal_id   | {proposalId}                   |