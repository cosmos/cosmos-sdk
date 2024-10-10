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
        amount: Item<Option<Coin>>,
        #[state]
        beneficiary: Item<AccountID>,
        #[state]
        unlock_time: Item<Time>,
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
                // return Err(UnlockError::FundsNotReceivedYet);
                todo!()
            }
            unsafe { ixc_core::account_api::self_destruct(ctx)?; }
            Ok(())
        }
    }



    #[handler_api]
    pub trait VestingAPI {
        fn unlock<'a>(&self, ctx: &'a mut Context, eb: &mut EventBus<UnlockEvent>) -> Result<(), UnlockError>;
    }

    #[derive(SchemaValue, Clone)]
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
        fn on_receive<'a>(&self, ctx: &mut Context<'a>, from: AccountID, denom: &str, amount: u128) -> Result<()>;
    }

    #[handler_api]
    #[automock]
    pub trait BlockInfoAPI {
        fn get_block_time<'a>(&self, ctx: &Context<'a>) -> Result<Time>;
    }

    #[publish]
    impl ReceiveHook for FixedVesting {
        fn on_receive<'a>(&self, ctx: &mut Context<'a>, from: AccountID, denom: &str, amount: u128) -> Result<()> {
            if ctx.caller() != self.bank_client.account_id() {
                bail!("only the bank can send funds to this account");
            }
            if let Some(_) = self.amount.get(ctx)? {
                bail!("already received deposit");
            }
            // Set the amount to unlock
            self.amount.set(ctx, Some(Coin{
                denom: denom.to_string(),
                amount,
            }))?;
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
    use ixc_core::account_api::ROOT_ACCOUNT;
    use ixc_core::handler::{Client, Service};
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
        bank_mock.expect_send().returning(|_, _, _| Ok(()));
        let bank_id = app.add_mock(&mut root, MockHandler::of::<dyn BankAPI>(Box::new(bank_mock))).unwrap();
        let mut bank_ctx = app.client_context_for(bank_id);

        // initialize block info
        let mut block_mock = MockBlockInfoAPI::new();
        // expect the unlock time to be 7 days from the default time
        let time0 = Time::default();
        block_mock.expect_get_block_time().returning(move |_| Ok(time0));
        let time1 = time0.add(Duration::DAY * 7);
        block_mock.expect_get_block_time().returning(move |_| Ok(time1));
        let block_id = app.add_mock(&mut root, MockHandler::of::<dyn BlockInfoAPI>(Box::new(block_mock))).unwrap();

        // initialize the vesting account
        let beneficiary = app.new_client_account().unwrap();
        let unlock_time = time0.add(Duration::DAY * 5);
        let vesting_acct = create_account(&mut root, FixedVestingCreate {
            beneficiary,
            unlock_time,
        }).unwrap();

        // pretend to be bank and deposit the initial funds
        let receive_hook_client = <dyn ReceiveHook>::new_client(vesting_acct.account_id());
        let funder_id = app.new_client_account().unwrap();
        receive_hook_client.on_receive(&mut bank_ctx, funder_id, "foo", 1000).unwrap();
        // expect that sending a second time returns an error
        let res = receive_hook_client.on_receive(&mut bank_ctx, funder_id, "foo", 1000);
        assert!(res.is_err());
    }
}

fn main() {}