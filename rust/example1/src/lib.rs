pub mod erc20;
mod bank;

pub struct VarChar<const N: usize>;
pub struct VarArray<T, const N: usize>;
pub struct Time(u64);
pub struct Item<T>;
type Result<T> = core::result::Result<T, String>;

type Denom = VarChar<256>;

#[derive(Struct)]
pub struct Coin {
    denom: Denom,
    amount: u128,
}

#[derive_client]
pub trait Bank {
    fn send(&self, ctx: &mut Context, from: Address, to: Address, amount: &[Coin]) -> Result<()>;
}

pub trait BlockService {
    fn current_time(&self, ctx: &Context) -> Result<Time>;
}

#[derive_client(Bank)]
pub struct BankClient;

#[derive_client(BlockServiceClient)]
pub struct BlockServiceClient;

#[derive(Account)]
pub struct FixedVestingAccount {
    beneficiary: Item<Address>,
    balance: Item<VarArray<Coin, 16>>,
    unlock_time: Item<Time>,
    bank_client: BankClient,
    block_service_client: BlockServiceClient,
}

impl FixedVestingAccount {
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