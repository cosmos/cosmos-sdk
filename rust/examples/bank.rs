#![allow(missing_docs)]
#[ixc::handler(Bank)]
pub mod bank {
    use ixc::*;

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
        fn create(&self, ctx: &mut Context, init_denom: &str, init_balance: u128) -> Result<()> {
            self.balances.add(ctx, (ctx.caller(), init_denom), init_balance)?;
            Ok(())
        }
    }

    #[publish]
    impl BankAPI for Bank {
        fn get_balance(&self, ctx: &Context, account: AccountID, denom: &str) -> Result<u128> {
            let amount = self.balances.get(ctx, (account, denom))?;
            Ok(amount)
        }

        fn send<'a>(&self, ctx: &'a mut Context, to: AccountID, amount: &[Coin<'a>], evt: EventBus<EventSend>) -> Result<()> {
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
    use ixc_core::handler::{Client, ClientFactory};
    use super::bank::*;
    use ixc_testing::*;

    #[test]
    fn test() {
        let mut app = TestApp::default();
        app.register_handler::<Bank>().unwrap();
        let mut alice = app.new_client_context().unwrap();
        let mut bob = app.new_client_context().unwrap();
        let bank_client = create_account(&mut alice, BankCreate { init_denom: "foo", init_balance: 1000 }).unwrap();
        let alice_balance = bank_client.get_balance(&alice, alice.account_id(), "foo").unwrap();
        assert_eq!(alice_balance, 1000);
        bank_client.send(&mut alice, bob.account_id(), &[Coin { denom: "foo", amount: 100 }]).unwrap();
        let alice_balance = bank_client.get_balance(&alice, alice.account_id(), "foo").unwrap();
        assert_eq!(alice_balance, 900);
        let bob_balance = bank_client.get_balance(&bob, bob.account_id(), "foo").unwrap();
        assert_eq!(bob_balance, 100);
    }
}

fn main() {}