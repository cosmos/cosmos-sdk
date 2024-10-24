#![allow(missing_docs)]
#[ixc::handler(Bank)]
pub mod bank {
    use mockall::automock;
    use ixc::*;
    use ixc_core::error::unimplemented_ok;
    use ixc_core::handler::Service;

    #[derive(Resources)]
    pub struct Bank {
        #[state(prefix = 1, key(address, denom), value(amount))]
        pub(crate) balances: AccumulatorMap<(AccountID, Str)>,
        #[state(prefix = 2, key(denom), value(total))]
        pub(crate) supply: AccumulatorMap<Str>,
        #[state(prefix = 3)]
        super_admin: Item<AccountID>,
        #[state(prefix = 4)]
        global_send_hook: Item<AccountID>,
        #[state(prefix = 5)]
        denom_admins: Map<Str, AccountID>,
        #[state(prefix = 6)]
        denom_send_hooks: Map<Str, AccountID>,
        #[state(prefix = 6)]
        denom_burn_hooks: Map<Str, AccountID>,
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
    #[automock]
    pub trait SendHook {
        fn on_send<'a>(&self, ctx: &mut Context<'a>, from: AccountID, to: AccountID, denom: &str, amount: u128) -> Result<()>;
    }

    #[handler_api]
    #[automock]
    pub trait BurnHook {
        fn on_burn<'a>(&self, ctx: &mut Context<'a>, from: AccountID, denom: &str, amount: u128) -> Result<()>;
    }

    #[handler_api]
    #[automock]
    pub trait ReceiveHook {
        fn on_receive<'a>(&self, ctx: &mut Context<'a>, from: AccountID, denom: &str, amount: u128) -> Result<()>;
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
        pub fn create(&self, ctx: &mut Context) -> Result<()> {
            self.super_admin.set(ctx, ctx.caller())?;
            Ok(())
        }

        #[publish]
        pub fn create_denom(&self, ctx: &mut Context, denom: &str, admin: AccountID) -> Result<()> {
            ensure!(self.super_admin.get(ctx)? == ctx.caller(), "not authorized");
            self.denom_admins.set(ctx, denom, admin)?;
            Ok(())
        }

        #[publish]
        pub fn set_global_send_hook(&self, ctx: &mut Context, hook: AccountID) -> Result<()> {
            ensure!(self.super_admin.get(ctx)? == ctx.caller(), "not authorized");
            self.global_send_hook.set(ctx, hook)?;
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
                unimplemented_ok(receive_hook.on_receive(ctx, from, coin.denom, coin.amount))?;
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
                .ok_or(error!("denom not defined"))?;
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
    use ixc_core::account_api::ROOT_ACCOUNT;
    use ixc_core::handler::{Client, Service};
    use ixc_core::routing::{find_route, Router};
    use ixc_message_api::code::ErrorCode;
    use ixc_message_api::handler::{Allocator, HostBackend, RawHandler};
    use ixc_message_api::packet::MessagePacket;
    use super::bank::*;
    use ixc_testing::*;

    #[test]
    fn test() {
        // initialize the app
        let mut app = TestApp::default();
        // register the Bank handler
        app.register_handler::<Bank>().unwrap();

        // create a new client context for the root account and initialize bank
        let mut root = app.client_context_for(ROOT_ACCOUNT);
        let bank_client = create_account::<Bank>(&mut root, BankCreate {}).unwrap();

        // register a mock global send hook to test that it is called
        let mut mock_global_send_hook = MockSendHook::new();
        // expect that the send hook is only called 1x in this test
        mock_global_send_hook.expect_on_send().times(1).returning(|_, _, _, _, _| Ok(()));
        let mut mock = MockHandler::new();
        mock.add_handler::<dyn SendHook>(Box::new(mock_global_send_hook));
        let mock_id = app.add_mock(mock).unwrap();
        bank_client.set_global_send_hook(&mut root, mock_id).unwrap();

        // alice gets to manage the "foo" denom and mints herself 1000 foo coins
        let mut alice = app.new_client_context().unwrap();
        let alice_id = alice.self_account_id();
        bank_client.create_denom(&mut root, "foo", alice_id).unwrap();
        bank_client.mint(&mut alice, alice_id, "foo", 1000).unwrap();

        // ensure alice has 1000 foo coins
        let alice_balance = bank_client.get_balance(&alice, alice_id, "foo").unwrap();
        assert_eq!(alice_balance, 1000);


        // alice sends 100 foo coins to bob
        let mut bob = app.new_client_context().unwrap();
        bank_client.send(&mut alice, bob.self_account_id(), &[Coin { denom: "foo", amount: 100 }]).unwrap();

        // ensure alice has 900 foo coins and bob has 100 foo coins
        let alice_balance = bank_client.get_balance(&alice, alice.self_account_id(), "foo").unwrap();
        assert_eq!(alice_balance, 900);
        let bob_balance = bank_client.get_balance(&bob, bob.self_account_id(), "foo").unwrap();
        assert_eq!(bob_balance, 100);

        // look inside bank to check the balance of alice directly as well as the supply of foo
        app.exec_in(&bank_client, |bank, ctx| {
            let alice_balance = bank.balances.get(ctx, (alice_id, "foo")).unwrap();
            assert_eq!(alice_balance, 900);
            let foo_supply = bank.supply.get(ctx, "foo").unwrap();
            assert_eq!(foo_supply, 1000);
        })
    }
}

fn main() {}