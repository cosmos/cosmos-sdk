pub mod fixed_size {
    use arrayvec::ArrayString;
    use crypto_bigint::U256;
    use cosmos_core::{Address, Context, Result};
    use cosmos_core_macros::service;

    #[service(solidity)]
    pub trait ERC20 {
        fn name(&self, ctx: &Context) -> Result<ArrayString<256>>;
        fn symbol(&self, ctx: &Context) -> Result<ArrayString<32>>;
        fn decimals(&self, ctx: &Context) -> Result<u8>;
        fn total_supply(&self, ctx: &Context) -> Result<U256>;
        fn balance_of(&self, ctx: &Context, owner: Address) -> Result<U256>;
        fn transfer(&self, ctx: &mut Context, to: Address, value: U256) -> Result<bool>;
        fn transfer_from(&self, ctx: &mut Context, from: Address, to: Address, value: U256) -> Result<bool>;
        fn approve(&self, ctx: &mut Context, spender: Address, value: U256) -> Result<bool>;
        fn allowance(&self, ctx: &Context, owner: Address, spender: Address) -> Result<U256>;
    }
}

pub mod dynamic_size {
    use crypto_bigint::U256;
    use cosmos_core::{Address, Context, Result};
    use cosmos_core_macros::service;

    #[service(solidity)]
    pub trait ERC20 {
        fn name(&self, ctx: &Context) -> Result<String>;
        fn symbol(&self, ctx: &Context) -> Result<String>;
        fn decimals(&self, ctx: &Context) -> Result<u8>;
        fn total_supply(&self, ctx: &Context) -> Result<U256>;
        fn balance_of(&self, ctx: &Context, owner: Address) -> Result<U256>;
        fn transfer(&self, ctx: &mut Context, to: Address, value: U256) -> Result<bool>;
        fn transfer_from(&self, ctx: &mut Context, from: Address, to: Address, value: U256) -> Result<bool>;
        fn approve(&self, ctx: &mut Context, spender: Address, value: U256) -> Result<bool>;
        fn allowance(&self, ctx: &Context, owner: Address, spender: Address) -> Result<U256>;
    }
}

mod example;