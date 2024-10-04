#![allow(missing_docs)]
#[ixc::handler(Bank)]
pub mod bank {
    use ixc::*;

    pub struct Bank {
        #[schema(name(address, denom), value(amount))]
        balances: Map<(AccountID, String), u128>,
    }

    #[derive(SchemaValue)]
    pub struct Coin<'a> {
        pub denom: &'a str,
        pub amount: u128,
    }

    #[handler_api]
    pub trait BankAPI {
        fn send(&self, ctx: &mut Context, to: AccountID, amount: &[Coin], evt: &mut EventBus<EventSend>) -> Result<()>;
    }

    #[derive(SchemaValue)]
    pub struct EventSend<'a> {
        pub from: Address,
        pub to: Address,
        pub coin: Coin<'a>,
    }

    impl BankAPI for Bank {
        fn send(&self, ctx: &mut Context, to: AccountID, amount: &[Coin], evt: &mut EventBus<EventSend>) -> Result<()> {
            for coin in amount {
                self.balances.safe_sub(ctx, (ctx.sender(), coin.denom), coin.amount)?;
                self.balances.add(ctx, (to, coin.denom), coin.amount)?;
                evt.emit(EventSend {
                    from: ctx.sender(),
                    to,
                    coin: coin.clone(),
                })?;
            }
            Ok(())
        }
    }
}

fn main() {}