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