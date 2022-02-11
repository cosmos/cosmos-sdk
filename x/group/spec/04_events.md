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

## EventCreateGroupPolicy

| Type                                         | Attribute Key | Attribute Value                              |
|----------------------------------------------|---------------|----------------------------------------------|
| message                                      | action        | /cosmos.group.v1beta1.Msg/CreateGroupPolicy |
| cosmos.group.v1beta1.EventCreateGroupPolicy | address       | {groupPolicyAddress}                        |

## EventUpdateGroupPolicy

| Type                                         | Attribute Key | Attribute Value                                                               |
|----------------------------------------------|---------------|-------------------------------------------------------------------------------|
| message                                      | action        | /cosmos.group.v1beta1.Msg/UpdateGroupPolicy{Admin\|Metadata\|DecisionPolicy} |
| cosmos.group.v1beta1.EventUpdateGroupPolicy | address       | {groupPolicyAddress}                                                         |

## EventCreateProposal

| Type                                     | Attribute Key | Attribute Value                          |
|------------------------------------------|---------------|------------------------------------------|
| message                                  | action        | /cosmos.group.v1beta1.Msg/CreateProposal |
| cosmos.group.v1beta1.EventCreateProposal | proposal_id   | {proposalId}                             |

## EventWithdrawProposal

| Type                                       | Attribute Key | Attribute Value                            |
|--------------------------------------------|---------------|--------------------------------------------|
| message                                    | action        | /cosmos.group.v1beta1.Msg/WithdrawProposal |
| cosmos.group.v1beta1.EventWithdrawProposal | proposal_id   | {proposalId}                               |

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
