#![allow(missing_docs)]
#[ixc::handler(SimpleAsset)]
pub mod simple_asset {
    use ixc::*;

    #[derive(Resources)]
    pub struct SimpleAsset {
        #[state]
        owner: Item<AccountID>,

        #[state(name(address), value(amount))]
        balances: Map<AccountID, u128>,
    }

    impl SimpleAsset {
        #[on_create]
        pub fn init(&self, ctx: &mut Context, initial_balance: u128) -> Result<()> {
            self.owner.set(ctx, ctx.caller())?;
            self.balances.set(ctx, ctx.caller(), initial_balance)
        }

        #[publish]
        pub fn get_balance(&self, ctx: &Context, account: AccountID) -> Result<u128> {
            let res = self.balances.get(ctx, &account)?;
            Ok(res.unwrap_or(0))
        }
    }
}

#[cfg(test)]
mod tests {
    use ixc_core::account_api::create_account;
    use ixc_core::handler::ClientFactory;
    use ixc_testing::*;
    use crate::simple_asset::SimpleAsset;

    #[test]
    fn test() {
        let mut app = TestApp::default();
        app.register_handler::<SimpleAsset>().unwrap();
        let mut alice = app.new_client_context().unwrap();
        let mut bob = app.new_client_context();
        // create_account(&mut alice, &SimpleAsset::init(100)).unwrap();
    }
}

fn main() {}
