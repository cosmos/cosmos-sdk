
# ADR 045: Admin module

## Changelog

- 2021/07/16: Initial draft

## Status

Proposed

## Abstract

This ADR proposes a new module - admin module, that adds simplified governance workflow with authority-based proposal acceptance. A new role is added - administrator. Admins will be capable of:
- submitting a proposal that will automatically be accepted;
- add/remove other admins;

## Context

The current proposal workflow is based on a concept of validators (token holders) who vote to accept/reject a proposal. The tallying of those votes is based on the voting power (stake-based voting) of validators and their delegators. This stake-based multi-step approach provides a high level of security necessary for mainnets.

However, the specifics of testnets - often updates with new features that need to be well tested.

Almost every feature and every scenario require some proposals to be accepted. For example, changing a test account's balance (thanks to blockchain technology) is not as easy as changing some number in a bank database. Any modification of a blockchain state requires a full proposal workflow with the majority of stakeholders to participate in voting. The situation is even more complicated in testnet - there's no motivation for users to participate in voting/governance. The full governance process is too long, so it slows down testing and the delivery of new features.

## Decision

We propose to create an admin module with simplified governance functionality. Admin module allows submitting proposals that are automatically accepted after verifying the Proposer ID, and managing the list of admins.

Admin module `Msg` and `Query` services will have the following interface:
```
type MsgService interface {
    // SubmitProposal defines a method to create new proposal with a given content.
    SubmitProposal(context.Context, *MsgSubmitProposal) (*MsgSenMsgSubmitProposalResponsedResponse, error)

    // AddAdmin defines a method to add a new admin
    AddAdmin(context.Context, *MsgAddAdmin) error

    // DeleteAdmin defines a method to delete an admin
    DeleteAdmin(context.Context, *MsgDeleteAdmin) error
}

type MsgSubmitProposal struct {
    Content        *types.Any

    //This field is generally unnecessary and kept only for similarity with the Gov module, can be removed
    InitialDeposit github_com_cosmos_cosmos_sdk_types.Coins

    Proposer       string
}

type MsgSubmitProposalResponse struct {
    ProposalId uint64
}

type MsgAddAdminRequest struct {
    Admin       string
    Requester string
}

type MsgDeleteAdminRequest struct {
    Admin        string
    Requester string
}

type QueryService interface {
    // ListAdmins defines a method to list admins
    ListAdmins(context.Context) (*QueryListAdminsResponse, error)
}

type QueryListAdminsResponse struct {
    AdminAddresses []string
}
```

### Submitting a Proposal

`MsgService.SubmitProposal` verifies the `MsgSubmitProposal.Proposer` to be in the list of admins. If so, the proposal is successfully submitted and added to the execution queue.

Admin module will have a separate `KVStore` with Addresses of the current admins' accounts.

Admin `KVStore` stores pairs in the following way:
```
{sdk.Address: true}
```

The Admin module `EndBlock` handler executes all proposals in the queue and calls logic in other modules (as the Governance module does).

Admin module will be able to exist alongside with running Governance module. They will manage two independent proposal queues, and each module will handle proposals only from its own queue.

The logic of proposal acceptance in `EndBlocker` handler
```
func EndBlocker {
    // logic to automatically accept any proposal from the admin module proposal queue and call state transitions in other modules similar to the Governance module https://github.com/cosmos/cosmos-sdk/blob/e17be874bb0e3a246b74752e9b8894855cab9b03/x/gov/abci.go#L58
}
```

### Adding a new admin
- First admin address should be in `genesis.json`.
- `MsgService.AddAdmin` handler verifies `MsgAddAdminRequest.Requester` to be in the list of admins. If so, it adds `MsgAddAdminRequest.Admin` to the list of admins.

### Removing an admin
`MsgService.DeleteAdmin` handler verifies `MsgDeleteAdminRequest.Requester` to be in the list of admins. If so, it deletes `MsgDeleteAdminRequest.Admin` from the list of admins.

### Turning the admin module off
There'll be two ways to switch off the admin module:
- the last admin deletes themself;
- the admin module is removed via software upgrade mechanism.
  
It will allow using the admin module during the first stages after a network launch and switcing it off later.

### Adding the admin module to existing testnets
Software upgrade mechanism allow adding the admins module to existing testnets via governance.

### Interaction with the admin module

There'll be new CLI commands for admin module functionality:
```sh
# submit proposal
simd tx admin submit-proposal --title="Test Proposal" --description="My awesome proposal" --type="Text" --deposit="10test" --from <mykey>
# adding new admin
simd tx admin add-admin [new_admin_address]  --from <mykey>
# removing admin
simd tx admin delete-admin [admin_address]  --from <mykey>
# query list of admins
simd query admin list
```

We also propose to create a simple web interface that demonstrates the work of the admin module. It will include the following features:
- review the current parameters of some existing modules;
- connect a wallet;
- submit a proposal;
- modify the list of admins.

### Positive
- Simplifies updating and applying new changes for the modules -> accelerates development and testing

### Negative
`None`

### Neutral
- Needs minor changes to simapp.
- Need changes to genesis.json (to add first admins).
