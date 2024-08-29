use arrayvec::ArrayString;
use cosmos_core::{Address, Context, Result, Item, Map, OnCreate};
use cosmos_core_macros::{Account, Serializable, State};
use crypto_bigint::U256;
use serde::Serialize;

#[derive(Account, State)]
pub struct BasicERC20 {
    #[item(prefix=0)]
    owner: Item<Address>,

    #[item(prefix=1)]
    name: Item<ArrayString<256>>,

    #[item(prefix=2)]
    symbol: Item<ArrayString<32>>,

    #[item(prefix=3)]
    decimals: Item<u8>,

    #[item(prefix=4)]
    supply: Item<U256>,

    #[map(prefix=5)]
    balances: Map<Address, U256>
}

#[derive(Serializable)]
pub struct Init {
    name: ArrayString<256>,
    symbol: ArrayString<32>,
    decimals: u8,
}

impl OnCreate for BasicERC20 {
    type InitMessage = Init;

    fn on_create(&self, ctx: &mut Context, msg: &Self::InitMessage) -> Result<()> {
        self.owner.set(ctx, &ctx.self_address())?;
        self.name.set(ctx, &msg.name)?;
        self.symbol.set(ctx, &msg.symbol)?;
        self.decimals.set(ctx, &msg.decimals)?;
        Ok(())
    }
}

impl crate::erc20::fixed_size::ERC20 for BasicERC20 {
    fn name(&self, ctx: &Context) -> Result<ArrayString<256>> {
        self.name.get(ctx)
    }

    fn symbol(&self, ctx: &Context) -> Result<ArrayString<32>> {
        self.symbol.get(ctx)
    }

    fn decimals(&self, ctx: &Context) -> Result<u8> {
        self.decimals.get(ctx)
    }

    fn total_supply(&self, ctx: &Context) -> Result<U256> {
        self.supply.get(ctx)
    }

    fn balance_of(&self, ctx: &Context, owner: Address) -> Result<U256> {
        self.balances.get(ctx, &owner)
    }

    fn transfer(&self, ctx: &mut Context, to: Address, value: U256) -> Result<bool> {
        self.balances.set(ctx, &to, &value)?;
        Ok(true)
    }

    fn transfer_from(&self, ctx: &mut Context, from: Address, to: Address, value: U256) -> Result<bool> {
        todo!()
    }

    fn approve(&self, ctx: &mut Context, spender: Address, value: U256) -> Result<bool> {
        todo!()
    }

    fn allowance(&self, ctx: &Context, owner: Address, spender: Address) -> Result<U256> {
        todo!()
    }
}

trait Mintable {
    fn mint(&self, ctx: &mut Context, to: Address, value: U256) -> Result<bool>;
}

impl Mintable for BasicERC20 {
    fn mint(&self, ctx: &mut Context, to: Address, value: U256) -> Result<bool> {
        if ctx.sender() != self.owner.get(ctx)? {
            return Err("only owner can mint".to_string());
        }

        let supply = self.supply.get(ctx)?;
        let new_supply = supply + value;
        self.supply.set(ctx, &new_supply)?;
        let balance = self.balances.get(ctx, &to)?; // TODO check if balance exists
        self.balances.set(ctx, &to, &(balance + value))?;
        Ok(true)
    }
}