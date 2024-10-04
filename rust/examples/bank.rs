#![allow(missing_docs)]
#[ixc::handler(Bank)]
pub mod bank {
    use ixc::*;
    use ixc_core::handler::ClientFactory;

    #[derive(Resources)]
    pub struct Bank {
        #[state(key(address, denom), value(amount))]
        balances: AccumulatorMap<(AccountID, Str)>,
    }

    #[derive(SchemaValue)]
    #[sealed]
    pub struct Coin<'a> {
        pub denom: &'a str,
        pub amount: u128,
    }

    #[handler_api]
    pub trait BankAPI {
        fn get_balance(&self, ctx: &Context, account: AccountID, denom: &str) -> Result<u128>;
        fn send<'a>(&self, ctx: &'a mut Context, to: AccountID, amount: &[Coin<'a>], evt: EventBus<EventSend<'_>>) -> Result<()>;
    }

    #[derive(SchemaValue)]
    #[non_exhaustive]
    pub struct EventSend<'a> {
        pub from: AccountID,
        pub to: AccountID,
        pub coin: Coin<'a>,
    }

    impl Bank {
        #[on_create]
        fn create(&self, ctx: &mut Context) -> Result<()> {
            Ok(())
        }
    }

    impl BankAPI for Bank {
        fn get_balance(&self, ctx: &Context, account: AccountID, denom: &str) -> Result<u128> {
            self.balances.get(ctx, (account, denom))
        }

        fn send(&self, ctx: &mut Context, to: AccountID, amount: &[Coin], evt: EventBus<EventSend>) -> Result<()> {
            for coin in amount {
                self.balances.safe_sub(ctx, (ctx.caller(), coin.denom), coin.amount)?;
                self.balances.add(ctx, (to, coin.denom), coin.amount)?;
                // evt.emit(EventSend {
                //     from: ctx.sender(),
                //     to,
                //     coin: coin.clone(),
                // })?;
            }
            Ok(())
        }
    }
}

#[cfg(test)]
mod tests {
    use super::bank::*;
    use ixc_testing::*;

    #[test]
    fn test() {
        let mut app = TestApp::default();
        app.register_handler::<Bank>().unwrap();
    }
}

fn main() {}