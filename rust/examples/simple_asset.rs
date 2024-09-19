#![allow(missing_docs)]
#[interchain_sdk::account_handler(SimpleAsset)]
pub mod simple_asset {
    use interchain_sdk::*;

    #[derive(Resources)]
    pub struct SimpleAsset {
        #[schema]
        owner: Item<Address>,

        #[schema(name(address), value(amount))]
        balances: UInt128Map<Address>,
    }

    impl SimpleAsset {
        #[on_create]
        pub fn init(&self, ctx: &mut Context, initial_balance: u128) -> Response<()> {
            // self.owner.set(ctx, ctx.caller().clone())?;
            // self.balances.add(ctx, ctx.caller(), initial_balance)
            todo!()
        }

        #[publish]
        pub fn get_balance(&self, ctx: &Context, address: Address) -> Response<u128> {
            // self.balances.get(ctx, address)
            todo!()
        }
    }
}

#[cfg(test)]
mod tests {
    use interchain_core_testing::*;
}

fn main() {}
