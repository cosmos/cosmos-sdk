use arrayvec::ArrayVec;
use cosmos_core::{Address, Context, Item, Time, Result};
use cosmos_core_macros::{service, Account};
use crate::bank::Coin;

#[derive(Account)]
#[]
pub struct FixedVestingAccount {
    #[item(prefix=1)]
    beneficiary: Item<Address>,
    #[item(prefix=2)]
    balance: Item<ArrayVec<Coin, 16>>,
    #[item(prefix=3)]
    unlock_time: Item<Time>,

    bank_client: BankClient,
    block_service_client: BlockServiceClient,
}

#[service]
pub trait VestingAccount {
    fn try_unlock(&mut self, ctx: &mut Context) -> Result<()>;
}

impl VestingAccount for FixedVestingAccount {
    fn try_unlock(&mut self, ctx: &mut Context) -> Result<()> {
        let now = self.block_service_client.current_time(&ctx)?;
        if now < self.unlock_time.get(&ctx)? {
            return Err("not yet unlocked".to_string());
        }
        self.bank_client.send(
            &ctx.self_address(),
            &self.beneficiary,
            &self.balance,
        )
    }
}
