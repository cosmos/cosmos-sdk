#![allow(missing_docs)]
#[ixc::handler(FixedVesting)]
mod vesting {
    use ixc::*;
    use mockall::automock;
    use num_enum::{IntoPrimitive, TryFromPrimitive};
    use thiserror::Error;
    use ixc_core::handler::{Client, Service};

    #[derive(Resources)]
    pub struct FixedVesting {
        #[state]
        pub(crate) amount: Item<Option<Coin>>,
        #[state]
        pub(crate) beneficiary: Item<AccountID>,
        #[state]
        pub(crate) unlock_time: Item<Time>,
        #[client(65536)]
        bank_client: <dyn BankAPI as Service>::Client,
        #[client(65537)]
        block_client: <dyn BlockInfoAPI as Service>::Client,
    }

    impl FixedVesting {
        #[on_create]
        fn create(&self, ctx: &mut Context, beneficiary: AccountID, unlock_time: Time) -> Result<()> {
            self.beneficiary.set(ctx, beneficiary)?;
            self.unlock_time.set(ctx, unlock_time)?;
            Ok(())
        }
    }

    #[publish]
    impl VestingAPI for FixedVesting {
        fn unlock(&self, ctx: &mut Context, eb: &mut EventBus<UnlockEvent>) -> Result<(), UnlockError> {
            if self.unlock_time.get(ctx)? > self.block_client.get_block_time(ctx)? {
                bail!(UnlockError::NotTimeYet);
            }
            if let Some(amount) = self.amount.get(ctx)? {
                let beneficiary = self.beneficiary.get(ctx)?;
                self.bank_client.send(ctx, beneficiary, &[amount.clone()])?;
                eb.emit(ctx, &UnlockEvent {
                    to: beneficiary.clone(),
                    amount,
                })?;
            } else {
                bail!(UnlockError::FundsNotReceivedYet);
            }
            unsafe { ixc_core::account_api::self_destruct(ctx)?; }
            Ok(())
        }
    }


    #[handler_api]
    pub trait VestingAPI {
        fn unlock<'a>(&self, ctx: &'a mut Context, eb: &mut EventBus<UnlockEvent>) -> Result<(), UnlockError>;
    }

    #[derive(SchemaValue, Clone, PartialEq, Debug)]
    #[sealed]
    pub struct Coin {
        pub denom: String,
        pub amount: u128,
    }

    #[handler_api]
    #[automock]
    pub trait BankAPI {
        fn send<'a>(&self, ctx: &mut Context<'a>, to: AccountID, amount: &[Coin]) -> Result<(), SendError>;
    }

    #[handler_api]
    pub trait ReceiveHook {
        fn on_receive<'a>(&self, ctx: &mut Context<'a>, from: AccountID, amount: &[Coin]) -> Result<()>;
    }

    #[handler_api]
    #[automock]
    pub trait BlockInfoAPI {
        fn get_block_time<'a>(&self, ctx: &Context<'a>) -> Result<Time>;
    }

    #[publish]
    impl ReceiveHook for FixedVesting {
        fn on_receive<'a>(&self, ctx: &mut Context<'a>, from: AccountID, amount: &[Coin]) -> Result<()> {
            if ctx.caller() != self.bank_client.account_id() {
                bail!("only the bank can send funds to this account");
            }
            if let Some(_) = self.amount.get(ctx)? {
                bail!("already received deposit");
            }
            if amount.len() != 1 {
                bail!("expected exactly one coin");
            }
            let coin = &amount[0];
            // Set the amount to unlock
            self.amount.set(ctx, Some(coin.clone()))?;
            Ok(())
        }
    }

    #[derive(SchemaValue)]
    #[non_exhaustive]
    pub struct UnlockEvent {
        pub to: AccountID,
        pub amount: Coin,
    }

    #[derive(Clone, Debug, IntoPrimitive, TryFromPrimitive, Error)]
    #[repr(u8)]
    pub enum UnlockError {
        #[error("the unlock time has not arrived yet")]
        NotTimeYet,

        #[error("the vesting account has not received any funds yet")]
        FundsNotReceivedYet,
    }

    #[derive(Clone, Debug, IntoPrimitive, TryFromPrimitive, Error)]
    #[repr(u8)]
    pub enum SendError {
        #[error("insufficient funds")]
        InsufficientFunds,

        #[error("send blocked")]
        SendBlocked,
    }
}

#[cfg(test)]
mod tests {
    use std::ops::{AddAssign, SubAssign};
    use std::sync::{Arc, RwLock};
    use ixc_core::account_api::ROOT_ACCOUNT;
    use ixc_core::handler::{Client, Service};
    use ixc_message_api::code::ErrorCode::{HandlerCode, SystemCode};
    use ixc_message_api::code::SystemCode::{AccountNotFound, HandlerNotFound};
    use ixc_testing::*;
    use simple_time::{Duration, Time};
    use super::vesting::*;

    #[test]
    fn test_unlock() {
        let mut app = TestApp::default();
        app.register_handler::<FixedVesting>().unwrap();
        let mut root = app.client_context_for(ROOT_ACCOUNT);

        // initialize block info
        let mut bank_mock = MockBankAPI::new();
        // expect send to be called once with the correct coins
        let coins = vec![Coin { denom: "foo".to_string(), amount: 1000 }];
        let expected_coins = coins.clone();
        bank_mock.expect_send().times(1).returning(move |_, _, coins| {
            assert_eq!(coins, expected_coins);
            Ok(())
        });
        let bank_id = app.add_mock(&mut root, MockHandler::of::<dyn BankAPI>(Box::new(bank_mock))).unwrap();
        let mut bank_ctx = app.client_context_for(bank_id);

        // initialize block info
        let mut block_mock = MockBlockInfoAPI::new();
        let cur_time = Arc::new(RwLock::new(Time::default()));
        let cur_time_copy = cur_time.clone();
        block_mock.expect_get_block_time().returning(move |_| Ok(cur_time_copy.read().unwrap().clone()));
        let block_id = app.add_mock(&mut root, MockHandler::of::<dyn BlockInfoAPI>(Box::new(block_mock))).unwrap();

        // initialize the vesting account
        let beneficiary = app.new_client_account().unwrap();
        let unlock_time = Time::default().add(Duration::DAY * 5);
        let vesting_acct = create_account::<FixedVesting>(&mut root, FixedVestingCreate {
            beneficiary,
            unlock_time,
        }).unwrap();

        // try to unlock before the initial deposit but after the unlock time (we're time traveling)
        cur_time.write().unwrap().add_assign(Duration::DAY * 6);
        let res = vesting_acct.unlock(&mut root);
        assert!(res.is_err());
        assert_eq!(res.unwrap_err().code, HandlerCode(UnlockError::FundsNotReceivedYet));

        // pretend to be bank and deposit the initial funds
        let receive_hook_client = <dyn ReceiveHook>::new_client(vesting_acct.account_id());
        let funder_id = app.new_client_account().unwrap();
        receive_hook_client.on_receive(&mut bank_ctx, funder_id, &coins).unwrap();
        // expect that sending a second time returns an error
        let res = receive_hook_client.on_receive(&mut bank_ctx, funder_id, &coins);
        assert!(res.is_err());

        // peek inside to make sure everything is setup correctly
        app.exec_in(&vesting_acct, |vesting, ctx| {
            let beneficiary = vesting.beneficiary.get(ctx).unwrap();
            assert_eq!(beneficiary, beneficiary);
            let unlock_time = vesting.unlock_time.get(ctx).unwrap();
            assert_eq!(unlock_time, unlock_time);
            let amount = vesting.amount.get(ctx).unwrap();
            assert_eq!(amount, Some(Coin { denom: "foo".to_string(), amount: 1000 }));
        });

        // try unlocking before the unlock time
        cur_time.write().unwrap().sub_assign(Duration::DAY * 6);
        let res = vesting_acct.unlock(&mut root);
        assert!(res.is_err());
        assert_eq!(res.unwrap_err().code, HandlerCode(UnlockError::NotTimeYet));
        // try unlocking after the unlock time
        cur_time.write().unwrap().add_assign(Duration::DAY * 6);
        vesting_acct.unlock(&mut root).unwrap();
        // TODO check for unlock event
        // since the unlock succeeded, if we try to unlock again we should get account not found,
        // because the account self-destructed
        let res = vesting_acct.unlock(&mut root);
        assert!(res.is_err());
        assert_eq!(res.unwrap_err().code, SystemCode(AccountNotFound));
    }
}

fn main() {}