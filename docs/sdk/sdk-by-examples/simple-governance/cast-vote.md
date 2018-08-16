## Cast a vote to an existing proposal

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