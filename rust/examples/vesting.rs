#![allow(missing_docs)]
#[ixc::handler(FixedVesting)]
mod vesting {
    use ixc::*;

    use thiserror::Error;
    use crate::vesting::bank_api::{OnReceive, Coin, BankAPI};
    use crate::vesting::block_api::BlockInfoAPI;
    use crate::vesting::vesting_api::{UnlockError, UnlockEvent, VestingAPI};

    #[derive(Resources)]
    pub struct FixedVesting {
        #[state]
        amount: Item<Option<Coin>>,
        #[state]
        beneficiary: Item<Address>,
        #[state]
        unlock_time: Item<Time>,
        #[client]
        bank_client : BankAPI::Client,
        #[client]
        block_client : BlockInfoAPI::Client,
    }

    #[publish]
    impl FixedVesting {
        fn on_create(&self, ctx: &mut Context, beneficiary: &Address, unlock_time: Time) -> Response<()> {
            self.beneficiary.set(ctx, beneficiary)?.map_err(|_| ())?;
            self.unlock_time.set(ctx, unlock_time)?.map_err(|_| ())?;
            Ok(())
        }
    }

    #[publish]
    impl VestingAPI for FixedVesting {
        fn unlock(&self, ctx: &mut Context, eb: &mut EventBus<UnlockEvent>) -> Response<(), UnlockError> {
            if self.unlock_time.get(ctx)? > self.block_client.get_block_time(ctx) {
                return Err(UnlockError::NotTimeYet);
            }
            if let Some(amount) = self.amount.get(ctx)? {
                let beneficiary = self.beneficiary.get(ctx)?;
                self.bank_client.send(ctx, beneficiary, &[amount])?;
                eb.emit(UnlockEvent {
                    to: beneficiary.clone(),
                    amount: amount.clone(),
                })?;
            } else {
                return Err(UnlockError::FundsNotReceivedYet);
            }
            unsafe { interchain_core::self_destruct::self_destruct(ctx) }
        }
    }

    #[publish]
    impl OnReceive for FixedVesting {
        fn on_receive(&self, ctx: &mut Context, from: Address, amount: Coin) -> Response<()> {
            if ctx.caller() != ctx.get_module_address::<dyn BankAPI>()? {
                // Only accept from bank
                return Err(());
            }

            if let Some(_) = self.amount.get(ctx)? {
                // Already received
                return Err(());
            }
            // Set the amount to unlock
            self.amount.set(ctx, Some(amount))?;
            Ok(())
        }
    }

    mod vesting_api {
        use ixc::*;
        use crate::vesting::bank_api::Coin;

        #[account_api]
        pub trait VestingAPI {
            fn unlock(&self, ctx: &mut Context, eb: &mut EventBus<UnlockEvent>) -> Response<(), UnlockError>;
        }

        #[derive(StructCodec)]
        pub struct UnlockEvent<'a> {
            pub to: Address,
            pub amount: Coin<'a>,
        }

        #[derive(EnumCodec, thiserror::Error)]
        pub enum UnlockError {
            #[error("the unlock time has not arrived yet")]
            NotTimeYet,

            #[error("the vesting account has not received any funds yet")]
            FundsNotReceivedYet,
        }
    }

    mod bank_api {
        use std::borrow::Cow;
        use ixc::*;

        #[derive(StructCodec, Clone)]
        pub struct Coin {
            pub denom: String,
            pub amount: u128,
        }

        #[module_api]
        pub trait BankAPI {
            fn send(&self, ctx: &mut Context, to: Address, amount: &[Coin]) -> Response<()>;
        }

        #[account_api]
        pub trait OnReceive {
            fn on_receive(&self, ctx: &mut Context, from: Address, amount: Coin) -> Response<(), SendError>;
        }

        #[derive(EnumCodec, thiserror::Error)]
        pub enum SendError {
            #[error("insufficient funds")]
            InsufficientFunds,

            #[error("send blocked")]
            SendBlocked
        }
    }

    mod block_api {
        use ixc::*;

        #[module_api]
        pub trait BlockInfoAPI {
            fn get_block_time(&self, ctx: &Context) -> Time;
        }
    }
}

fn main() {}