# Setup And Design

## Get started

To get started, you just have to follow these simple steps:

1. Clone the [Cosmos-SDK](https://github.com/cosmos/cosmos-sdk/tree/develop)repo
2. Code the modules needed by your application that do not already exist.
3. Create your app directory. In the app main file, import the module you need and instantiate the different stores.
4. Launch your blockchain.

## Setup

### Prerequisites

- Have [go](https://golang.org/dl/) and [git](https://git-scm.com/downloads) installed
- Don't forget to set your `PATH` and `GOPATH`

### Setup work environment

Go to the [Cosmos-SDK repo](https://githum.com/cosmos/cosmos-sdk) and fork it. Then open a terminal and:

```bash
cd $GOPATH/src/github.com/your_username
git clone github.com/your_username/cosmos-sdk
cd cosmos-sdk
```

Now we'll add the origin Cosmos-SDK as upstream in case some cool feature or module gets merged:

```bash
git remote add upstream github.com/cosmos/cosmos-sdk
git fetch upstream
git rebase upstream/master
```

We will also create a branch dedicated to our module:

```bash
git checkout -b my_new_application
```

We are all set!

## Application design

### Application description

For this tutorial, we will code a **simple governance application**, accompagnied by a **simple governance module**. It will allow us to explain most of the basic notions required to build a functioning application on the Cosmos-SDK. Note that this is not the governance module used for the Cosmos Hub. A much more [advanced governance module](https://github.com/cosmos/cosmos-sdk/tree/develop/x/gov) will be used instead.

All the code for the `simple_governance` application can be found [here](https://github.com/gamarin2/cosmos-sdk/tree/module_tutorial/examples/simpleGov/x/simple_governance). You'll notice that the module and app aren't located at the root level of the repo but in the examples directory. This is just for convenience, you can code your module and application directly in the root directory.

Without further talk, let's get into it!

### Requirements

We will start by writting down your module's requirements. We are designing a simple governance module, in which we want:

- Simple text proposals, that any coin holder can submit.
- Proposals must be submitted with a deposit in Atoms. If the deposit is larger than a  `MinDeposit`, the associated proposal enters the voting period. Otherwise it is rejected. 
- Bonded Atom holders can vote on proposal on a 1 bonded Atom 1 vote basis
- Bonded Atom holders can choose between 3 options when casting a vote: `Yes`, `No` and `Abstain`.
- If, at the end of the voting period, there are more `Yes` votes than `No` votes, the proposal is accepted. Otherwise, it is rejected.
- Voting period is 2 weeks

When designing a module, it is good to adopt a certain methodology. Remember that a blockchain application is just a replicated state-machine. The state is the representation of the application at a given time. It is up to the application developer to define what the state represents, depending on the goal of the application. For example, the state of a simple cryptocurrency application will be a mapping of addresses to balances.

The state can be updated according to predefined rules. Given a state and a transaction, the state-machine (i.e. the application) will return a new state. In a blockchain application, transactions are bundled in blocks, but the logic is the same. Given a state and a set of transactions (a block), the application returns a new state. A SDK-module is just a subset of the application, but it is based on the same principles. As a result, module developers only have to define a subset of the state and a subset of the transaction types, which trigger state transitions.

In summary, we have to define:

- A `State`, which represents a subset of the current state of the application.
- `Transactions`, which contain messages that trigger state transitions.

### State

Here, we will define the types we need (excluding transaction types), as well as the stores in the multistore.

Our voting module is very simple, we only need a single type: `Proposal`. `Proposals` are item to be voted upon. They can be submitted by any user. A deposit has to be provided.

```go
type Proposal struct {
    Title           string          // Title of the proposal
    Description     string          // Description of the proposal
    Submitter       sdk.Address     // Address of the submitter. Needed to refund deposit if proposal is accepted.
    SubmitBlock     int64           // Block at which proposal is submitted. Also the block at which voting period begins.
    State           string          // State can be either "Open", "Accepted" or "Rejected"

    YesVotes        int64           // Total number of Yes votes
    NoVotes         int64           // Total number of No votes
    AbstainVotes    int64           // Total number of Abstain votes
}
```

In terms of store, we will just create one [KVStore](#kvstore) in the multistore to store `Proposals`. We will also store the `Vote` (`Yes`, `No` or `Abstain`) chosen by each voter on each proposal.


### Messages

As a module developer, what you have to define are not `Transactions`, but `Messages`. Both transactions and messages exist in the Cosmos-SDK, but a transaction differs from a message in that a message is contained in a transaction. Transactions wrap around messages and add standard information like signatures and fees. As a module developer, you do not have to worry about transactions, only messages.

Let us define the messages we need in order to modify the state. Based on the requirements above, we need to define two types of messages: 

- `SubmitProposalMsg`: to submit proposals
- `VoteMsg`: to vote on proposals

```go
type SubmitProposalMsg struct {
    Title           string      // Title of the proposal
    Description     string      // Description of the proposal
    Deposit         sdk.Coins   // Deposit paid by submitter. Must be > MinDeposit to enter voting period
    Submitter       sdk.Address // Address of the submitter
}
```

```go
type VoteMsg struct {
    ProposalID  int64           // ID of the proposal
    Option      string          // Option chosen by voter
    Voter       sdk.Address     // Address of the voter
}
```