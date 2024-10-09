#![allow(missing_docs)]
#[ixc::handler(Bank)]
pub mod bank {
    use ixc::*;
    use ixc_core::handler::ClientFactory;
    use ixc_message_api::code::{ErrorCode};

    #[derive(Resources)]
    pub struct Bank {
        #[state(prefix = 1, key(address, denom), value(amount))]
        balances: AccumulatorMap<(AccountID, Str)>,
        #[state(prefix = 2, key(denom), value(total))]
        supply: AccumulatorMap<Str>,
        #[state(prefix = 3)]
        super_admin: Item<AccountID>,
        #[state(prefix = 4)]
        denom_admins: Map<Str, AccountID>,
        #[state(prefix = 5)]
        denom_send_hooks: Map<Str, AccountID>,
        #[state(prefix = 6)]
        global_send_hook: Item<AccountID>
    }

    #[derive(SchemaValue, Clone)]
    #[sealed]
    pub struct Coin<'a> {
        pub denom: &'a str,
        pub amount: u128,
    }

    #[handler_api]
    pub trait BankAPI {
        fn get_balance(&self, ctx: &Context, account: AccountID, denom: &str) -> Result<u128>;
        fn send<'a>(&self, ctx: &'a mut Context, to: AccountID, amount: &[Coin<'a>], evt: EventBus<EventSend<'_>>) -> Result<()>;
        fn mint(&self, ctx: &mut Context, to: AccountID, denom: &str, amount: u128, evt: EventBus<EventMint<'_>>) -> Result<()>;
        fn burn(&self, ctx: &mut Context, from: AccountID, denom: &str, amount: u128, evt: EventBus<EventBurn<'_>>) -> Result<()>;
    }

    #[handler_api]
    pub trait SendHook {
        fn on_send(&self, ctx: &mut Context, from: AccountID, to: AccountID, denom: &str, amount: u128) -> Result<()>;
    }

    #[handler_api]
    pub trait ReceiveHook {
        fn on_receive(&self, ctx: &mut Context, from: AccountID, denom: &str, amount: u128) -> Result<()>;
    }

    #[derive(SchemaValue)]
    #[non_exhaustive]
    pub struct EventSend<'a> {
        pub from: AccountID,
        pub to: AccountID,
        pub coin: Coin<'a>,
    }

    #[derive(SchemaValue)]
    #[non_exhaustive]
    pub struct EventMint<'a> {
        pub to: AccountID,
        pub coin: Coin<'a>,
    }

    #[derive(SchemaValue)]
    #[non_exhaustive]
    pub struct EventBurn<'a> {
        pub from: AccountID,
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

        fn send<'a>(&self, ctx: &'a mut Context, to: AccountID, amount: &[Coin<'a>], mut evt: EventBus<EventSend<'a>>) -> Result<()> {
            let global_send = self.global_send_hook.get(ctx)?;
            for coin in amount {
                if !global_send.is_empty() {
                    let hook_client = <dyn SendHook>::new_client(global_send);
                    hook_client.on_send(ctx, ctx.caller(), to, coin.denom, coin.amount)?;
                }
                if let Some(hook) = self.denom_send_hooks.get(ctx, coin.denom)? {
                    let hook_client = <dyn SendHook>::new_client(hook);
                    hook_client.on_send(ctx, ctx.caller(), to, coin.denom, coin.amount)?;
                }
                let from = ctx.caller();
                let receive_hook = <dyn ReceiveHook>::new_client(to);
                match receive_hook.on_receive(ctx, from, coin.denom, coin.amount) {
                    Ok(_) => {}
                    Err(e) => {
                        match e.code {
                            ErrorCode::SystemCode(ixc_message_api::code::SystemCode::MessageNotHandled) => {}
                            _ => bail!("receive blocked: {:?}", e),
                        }
                    }
                }
                self.balances.safe_sub(ctx, (from, coin.denom), coin.amount)?;
                self.balances.add(ctx, (to, coin.denom), coin.amount)?;
                evt.emit(ctx, &EventSend {
                    from,
                    to,
                    coin: coin.clone(),
                })?;
            }
            Ok(())
        }

        fn mint<'a>(&self, ctx: &mut Context, to: AccountID, denom: &'a str, amount: u128, mut evt: EventBus<EventMint<'a>>) -> Result<()> {
            let admin = self.denom_admins.get(ctx, denom)?
                .ok_or(fmt_error!("denom not defined"))?;
            ensure!(admin == ctx.caller(), "not authorized");
            self.supply.add(ctx, denom, amount)?;
            self.balances.add(ctx, (to, denom), amount)?;
            evt.emit(ctx, &EventMint {
                to,
                coin: Coin { denom, amount },
            })?;
            Ok(())
        }

        fn burn(&self, ctx: &mut Context, from: AccountID, denom: &str, amount: u128, mut evt: EventBus<EventBurn<'_>>) -> Result<()> {
            // TODO burn hooks
            todo!()
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