# Running The Application

## Installation

Once you have finallized your application, install it using `go get`. The following commands will install the pre-built modules and examples of the SDK as well as your `simpleGov` application:

```bash
go get github.com/<your_username>/cosmos-sdk
cd $GOPATH/src/github.com/<your_username>/cosmos-sdk
make get_vendor_deps
make install
make install_examples
```

Check that the app is correctly installed by typing:

```bash
simplegovcli -h
simplegovd -h
```

## Submit a proposal

Uuse the CLI to create a new proposal:

```bash
simplegovcli propose --title="Voting Period update" --description="Should we change the proposal voting period to 3 weeks?" --deposit=300Atoms
```

Or, via a json file:

```bash
simplegovcli propose --proposal="path/to/proposal.json"
```

Where proposal.json contains:

```json
{
  "title": "Voting Period Update",
  "description": "Should we change the proposal voting period to 3 weeks?",
  "type": "Text",
  "deposit": "300Atoms"
}
```

Get the details of your newly created proposal:

```bash
simplegovcli proposal 1
```

You can also check all the existing open proposals:

```bash
simplegovcli proposals --active=true
```

## Cast a vote 

Let's cast a vote on the created proposal:

```bash
simplegovcli vote --proposal-id=1 --option="No"
```

Get the value of the option from your casted vote :

```bash
simplegovcli proposal-vote 1 <your_address>
```

You can also check all the casted votes of a proposal:

```bash
simplegovcli proposals-votes 1
```