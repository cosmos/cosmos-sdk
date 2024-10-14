---
sidebar_position: 1
---

# Guide to On-Chain Multisig transactions

## Overview

Multisignature **on-chain** accounts are an improvement over the previous implementation as these introduce a new set of
features.

### Threshold and quorums

The previous implementation only allowed for m-of-n multisig accounts, where m is the number of signatures required to
authorize a transaction and n is the total number of signers. The new implementation allows for more flexibility by
introducing threshold and quorum values. The quorum is the minimum voting power to make a proposal valid, while the
threshol is the minimum of voting power of YES votes to pass a proposal.

### Revote

Multisigs can allow members to change their votes after the initial vote. This is useful when a member changes their mind
or when new information becomes available.

### Early execution

Multisigs can be configured to allow for early execution of proposals. This is useful when a proposal is time-sensitive or
when the proposer wants to execute the proposal as soon as it reaches the threshold. It can also be used to mimic the
behavior of the previous multisig implementation.

### Voting period

Multisigs can be configured to have a voting period. This is the time window during which members can vote on a proposal.
If the proposal does not reach the threshold within the voting period, it is considered failed.

## Setup

We'll create a multisig with 3 members with a 2/3 passing threshold.

First create the 3 members, Alice, Bob and Carol:

```bash!
simd keys add alice --keyring-backend test --home ./.testnets/node0/simd/
simd keys add bob --keyring-backend test --home ./.testnets/node0/simd/
simd keys add carol --keyring-backend test --home ./.testnets/node0/simd/
```

And we initialize them with some tokens (sent from one of our nodes):

```bash!
simd tx bank send $(simd keys show node0 --address  --keyring-backend=test --home ./.testnets/node0/simd/) $(simd keys show alice --address  --keyring-backend=test --home ./.testnets/node0/simd/) 100stake --fees 5stake --chain-id $CHAINID --keyring-backend test --home ./.testnets/node0/simd/
simd tx bank send $(simd keys show node0 --address  --keyring-backend=test --home ./.testnets/node0/simd/) $(simd keys show bob --address  --keyring-backend=test --home ./.testnets/node0/simd/) 100stake --fees 5stake --chain-id $CHAINID --keyring-backend test --home ./.testnets/node0/simd/
simd tx bank send $(simd keys show node0 --address  --keyring-backend=test --home ./.testnets/node0/simd/) $(simd keys show carol --address  --keyring-backend=test --home ./.testnets/node0/simd/) 100stake --fees 5stake --chain-id $CHAINID --keyring-backend test --home ./.testnets/node0/simd/
```

Now we craft our initialization message, in it we'll include the members' addresses, their weights and the configuration of our multisig.

```json
{
  "members": [
    {
      "address": "cosmos1pr26h2vq9adq3acvh37pz6wtk65u3y8798scq0",
      "weight": 1000
    },
    {
      "address": "cosmos1j4p2xlg393rg4mma0058alzgvkrjdddd2f5fll",
      "weight": 1000
    },
    {
      "address": "cosmos1vaqh39cdex9sgr46ef0tdln5cn0hdyd3s0lx4l",
      "weight": 1000
    }
  ],
  "config": {
    "threshold": 2000,
    "quorum": 2000,
    "voting_period": 86400,
    "revote": false,
    "early_execution": true
  }
}
```

In the configuration we set the threshold and quorum to the same, 2/3 of the members must vote yes to pass the proposal. Other configurations can set the quorum and threshold to different values to mimic how organizations work.

We've also set `early_execution` to true, to allow executing as soon as the proposal passes.

Voting period is in seconds, so we've set that to 24h. And finally `revote` was set to false, because we don't want to allow members to change their vote mid-way through.

To initialize the multisig, we have to run the `accounts init` passing the account type and the json we created.


```bash!
initcontents=$(cat init.json)
simd tx accounts init multisig $initcontents  --fees 5stake --chain-id $CHAINID --keyring-backend test --home ./.testnets/node0/simd/ --from alice
```

If everything goes well, we'll get back a tx hash, and we'll check the tx result to get our newly created multisig's address.

```bash!
simd q tx 472B5B4E181D2F399C0ACE4DEEB26FE4351D13E593ED8E793B005C48BFD32621 --output json | jq -r '.events[] | select(.type == "account_creation") | .attributes[] | select(.key == "address") | .value'
```

In this case, the address is `cosmos1uds6tz96dxfllz7tz3s3tm8tlg6x95g0mc2987sx6psjz98qlpss89sheu`. We can now send tokens to it, just like to a normal account.

```bash!
simd tx bank send $(simd keys show node0 --address  --keyring-backend=test --home ./.testnets/node0/simd/) cosmos1uds6tz96dxfllz7tz3s3tm8tlg6x95g0mc2987sx6psjz98qlpss89sheu 10000stake --fees 5stake --chain-id $CHAINID --keyring-backend test --home ./.testnets/node0/simd/
```

## Proposals

#### Create proposal

In this multisig, every action is a proposal. We'll do a simple proposal to send tokens from the multisig to Alice.

```json
{
  "proposal": {
    "title": "Send 1000 tokens to Alice",
    "summary": "Alice is a great multisig member so let's pay her.",
    "messages": [
      {
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from_address": "cosmos1uds6tz96dxfllz7tz3s3tm8tlg6x95g0mc2987sx6psjz98qlpss89sheu",
        "to_address": "cosmos1pr26h2vq9adq3acvh37pz6wtk65u3y8798scq0",
        "amount": [
          {
            "denom": "stake",
            "amount": "1000"
          }
        ]
      }
    ]
  }
}
```

> The content of messages was created using a simple `tx send` command and passing the flag `--generate-only` so we could copy the message.

Now we send the tx that will create the proposal:

```bash!
propcontents=$(cat createprop.json)
simd tx accounts execute cosmos1uds6tz96dxfllz7tz3s3tm8tlg6x95g0mc2987sx6psjz98qlpss89sheu cosmos.accounts.defaults.multisig.v1.MsgCreateProposal $propcontents --fees 5stake --chain-id $CHAINID --keyring-backend test --home ./.testnets/node0/simd/ --from alice
```

This will again return a tx hash that we can use to find out the newly created proposal.

```bash!
simd q tx 5CA4420B67FB040B3DF2484CB875E030123662F43AE9958A9F8028C1281C8654 --output json | jq -r '.events[] | select(.type == "proposal_created") | .attributes[] | select(.key == "proposal_id") | .value'
```

In this case, because this is the first proposal, we'll get that the proposal ID is 0. We can use this to query it.

```bash!
simd q accounts query cosmos1uds6tz96dxfllz7tz3s3tm8tlg6x95g0mc2987sx6psjz98qlpss89sheu cosmos.accounts.defaults.multisig.v1.QueryProposal '{"proposal_id":1}' 
```

We get back all the details from the proposal, including the end of the voting period and the current status of the proposal.

```yaml
response:
  '@type': /cosmos.accounts.defaults.multisig.v1.QueryProposalResponse
  proposal:
    messages:
    - '@type': /cosmos.bank.v1beta1.MsgSend
      amount:
      - amount: "1000"
        denom: stake
      from_address: cosmos1uds6tz96dxfllz7tz3s3tm8tlg6x95g0mc2987sx6psjz98qlpss89sheu
      to_address: cosmos1pr26h2vq9adq3acvh37pz6wtk65u3y8798scq0
    status: PROPOSAL_STATUS_VOTING_PERIOD
    summary: Alice is a great multisig member so let's pay her.
    title: Send 1000 tokens to Alice
    voting_period_end: "1717064354"
```

### Vote on the proposal

Just like before, we'll use `tx accounts execute`, but this time to vote. As we have a 2/3 passing threshold, we have to vote with at least 2 members.

```bash!
simd tx accounts execute cosmos1uds6tz96dxfllz7tz3s3tm8tlg6x95g0mc2987sx6psjz98qlpss89sheu cosmos.accounts.defaults.multisig.v1.MsgVote '{"proposal_id":0, "vote":"VOTE_OPTION_YES"}' --fees 5stake --chain-id $CHAINID --keyring-backend test --home ./.testnets/node0/simd/ --from alice --yes
simd tx accounts execute cosmos1uds6tz96dxfllz7tz3s3tm8tlg6x95g0mc2987sx6psjz98qlpss89sheu cosmos.accounts.defaults.multisig.v1.MsgVote '{"proposal_id":0, "vote":"VOTE_OPTION_YES"}' --fees 5stake --chain-id $CHAINID --keyring-backend test --home ./.testnets/node0/simd/ --from bob --yes
```

### Execute the proposal

Once we got enough votes, we can execute the proposal.

```bash!
simd tx accounts execute cosmos1uds6tz96dxfllz7tz3s3tm8tlg6x95g0mc2987sx6psjz98qlpss89sheu cosmos.accounts.defaults.multisig.v1.MsgExecuteProposal '{"proposal_id":0}' --fees 5stake --chain-id $CHAINID --keyring-backend test --home ./.testnets/node0/simd/ --from bob --yes
```

Querying the tx hash will get us information about the success or failure of the proposal execution.

```yaml
- attributes:
  - index: true
    key: proposal_id
    value: "0"
  - index: true
    key: yes_votes
    value: "2000"
  - index: true
    key: no_votes
    value: "0"
  - index: true
    key: abstain_votes
    value: "0"
  - index: true
    key: status
    value: PROPOSAL_STATUS_PASSED
  - index: true
    key: reject_err
    value: <nil>
  - index: true
    key: exec_err
    value: <nil>
  - index: true
    key: msg_index
    value: "0"
  type: proposal_tally
```

Now checking the multisig and Alice's balance, we'll see that the send was performed correctly.

```bash!
simd q bank balances cosmos1uds6tz96dxfllz7tz3s3tm8tlg6x95g0mc2987sx6psjz98qlpss89sheu
                                         
balances:
- amount: "9000"
  denom: stake
pagination:
  total: "1"
```

```bash!
simd q bank balances $(./build/simd keys show alice --address)                                              

balances:
- amount: "1080"
  denom: stake
pagination:
  total: "1"
```



