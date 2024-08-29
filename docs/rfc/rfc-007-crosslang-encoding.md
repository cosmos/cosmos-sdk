# RFC 007: Cross-Language Message Encoding

## Changelog

* 2024-08-23: Initial draft

## Background

> The next section is the "Background" section. This section should be at least two paragraphs and can take up to a whole
> page in some cases. The guiding goal of the background section is: as a newcomer to this project (new employee, team
> transfer), can I read the background section and follow any links to get the full context of why this change is  
> necessary?
>
> If you can't show a random engineer the background section and have them acquire nearly full context on the necessity
> for the RFC, then the background section is not full enough. To help achieve this, link to prior RFCs, discussions, and
> more here as necessary to provide context so you don't have to simply repeat yourself.


## Design Alternatives

_This section is non-standard and replaces the standard Proposal section._

### Rust APIs

#### Prost 

This would

```rust
#[derive(Clone, Debug, PartialEq, Message)]
pub struct MsgSend {
    from: String,
    to: String,
    amount: Vec<Coin>,    
}

#[derive(Clone, Debug, PartialEq, Message)]
pub struct Coin {
    denom: String,
    amount: String,
}

impl Bank {
    fn send(&self, ctx: &mut Context, &msg: MsgSend) -> Result<()> { /* ... */ }
}

fn example_send(bank: &Bank, ctx: &mut Context) -> Result<()> {
    let msg = MsgSend {
        from: "alice".to_string(), // allocates
        to: "bob".to_string(), // allocates
        coins: vec![Coin { amount: "100".to_string(), denom: "uatom".to_string() }] // allocates 3 times
    };
    bank.send(&mut ctx, &msg)
}
```

### ZeroPB

```rust
#[derive(ZeroPB)]
pub struct MsgSend {
    from: zeropb::Str,
    to: zeropb::Str,
    coins: zeropb::Repeated<Coin>
}

#[derive(ZeroPB)]
pub struct Coin {
    denom: zeropb::Str,
    amount: zeropb::Str,
}

impl Bank {
    fn send(&self, ctx: &mut Context, &msg: MsgSend) -> Result<()> { /* ... */ }
}

fn example_send(bank: &Bank, ctx: &mut Context) -> Result<()> {
    let mut msg = zerop::Root::<Bank>::new(); // one allocation only here
    msg.from.set("alice")?;
    msg.to.set("bob")?;    
    let mut amount_writer = msg.amount.start_write()?;
    let mut coin = amount_writer.append()?;
    coin.amount.set("100")?;
    coin.denom.set("uatom")?;
    bank.send(&mut ctx, &msg)
}
```

### BorrowPB

```rust
type Denom = VarChar<64>;

#[derive(Clone, Debug, PartialEq, Message)]
pub struct MsgSend<'a> {
    from: Address,
    to: Address,
    coins: Repeated<'a, Coin<'a>, 16>, // note the use of a fixed size Repeated buffer here holding up to 16 coins
}

#[derive(Clone, Debug, PartialEq, Message)]
pub struct Coin<'a> {
    denom: Denom,
    amount: u128,
}

impl Bank {
    fn send(&self, ctx: &mut Context, msg: &'a MsgSend<'a>) -> Result<()> { /* ... */ }
}

fn example_send(bank: &Bank, ctx: &mut Context) -> Result<()> {
    let mut coins = Repeated::<_, 16>::new(); // stack allocated
    coins.push(Coin{ amount: 100, denom: Varchar::new("uatom") }); // stack allocated
    let msg = MsgSend {
        from: Address::from_str("alice"), // using the stack here - fixed size buffer 
        to: Address::from_str("bob"), // using the stack here
        coins,
    }; // this whole struct gets allocated on the stack
    bank.send(&mut ctx, &msg)
}
```

### Regular Function Arguments

```rust
type Denom = VarChar<64>;

struct Coin {
    denom: Denom,
    amount: u128,
}

impl Bank {
    fn send(ctx: &Context, from: &Address, to: &Address, coins: &[Coin]) -> Result<()> {  /* ... */ }
}

fn example_send(bank: &Bank, ctx: &mut Context) -> Result<()> {
    let mut coins = ArrayVec::<_, 16>::new(); // stack allocated
    coins.push(Coin{ amount: 100, denom: Varchar::new("uatom") }); // stack allocated
    bank.send(&mut ctx, "alice", "bob", &coins);
}
```

## Abandoned Ideas (Optional)

> As RFCs evolve, it is common that there are ideas that are abandoned. Rather than simply deleting them from the
> document, you should try to organize them into sections that make it clear they're abandoned while explaining why they
> were abandoned.
>
> When sharing your RFC with others or having someone look back on your RFC in the future, it is common to walk the same
> path and fall into the same pitfalls that we've since matured from. Abandoned ideas are a way to recognize that path
> and explain the pitfalls and why they were abandoned.

## Decision

> This section describes alternative designs to the chosen design. This section
> is important and if an adr does not have any alternatives then it should be
> considered that the ADR was not thought through.

## Consequences (optional)

> This section describes the resulting context, after applying the decision. All
> consequences should be listed here, not just the "positive" ones. A particular
> decision may have positive, negative, and neutral consequences, but all of them
> affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section
> describing these incompatibilities and their severity. The ADR must explain
> how the author proposes to deal with these incompatibilities. ADR submissions
> without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

> {positive consequences}

### Negative

> {negative consequences}

### Neutral

> {neutral consequences}



### References

> Links to external materials needed to follow the discussion may be added here.
>
> In addition, if the discussion in a request for comments leads to any design
> decisions, it may be helpful to add links to the ADR documents here after the
> discussion has settled.

## Discussion

> This section contains the core of the discussion.
>
> There is no fixed format for this section, but ideally changes to this
> section should be updated before merging to reflect any discussion that took
> place on the PR that made those changes.
