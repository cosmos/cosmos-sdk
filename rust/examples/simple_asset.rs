#![allow(missing_docs)]
#[ixc::handler(SimpleAsset)]
pub mod simple_asset {
    use ixc::*;

    #[derive(Resources)]
    pub struct SimpleAsset {
        #[state]
        owner: Item<AccountID>,

        #[state]
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
            let res = self.balances.get(ctx, account)?;
            Ok(res.unwrap_or(0))
        }

        #[publish]
        pub fn send(&self, ctx: &mut Context, amount: u128, to: AccountID) -> Result<()> {
            let from = ctx.caller();
            let from_balance = self.balances.get(ctx, from)?.unwrap_or(0);
            if from_balance < amount {
                return Err(())
            }
            let to_balance = self.balances.get(ctx, to)?.unwrap_or(0);
            self.balances.set(ctx, from, from_balance - amount)?;
            self.balances.set(ctx, to, to_balance + amount)?;
            Ok(())
        }
    }
}

#[cfg(test)]
mod tests {
    use ixc_core::account_api::create_account;
    use ixc_core::handler::ClientFactory;
    use ixc_testing::*;
    use crate::simple_asset::{SimpleAsset, SimpleAssetInitMsg};

    #[test]
    fn test() {
        let mut app = TestApp::default();
        app.register_handler::<SimpleAsset>().unwrap();
        let mut alice = app.new_client_context().unwrap();
        let mut bob = app.new_client_context();
        create_account(&mut alice, SimpleAssetInitMsg{ initial_balance: 100 }).unwrap();
    }
}

fn main() {}
