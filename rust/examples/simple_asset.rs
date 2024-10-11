#![allow(missing_docs)]
#[ixc::handler(SimpleAsset)]
pub mod simple_asset {
    use ixc::*;

    #[derive(Resources)]
    pub struct SimpleAsset {
        #[state]
        owner: Item<AccountID>,

        #[state(key(account), value(balance))]
        balances: Map<AccountID, u128>,
    }

    impl SimpleAsset {
        #[on_create]
        pub fn init(&self, ctx: &mut Context, initial_balance: u128) -> Result<()> {
            let owner = ctx.caller();
            self.owner.set(ctx, &owner)?;
            self.balances.set(ctx, owner, initial_balance)?;
            Ok(())
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
            ensure!(from_balance >= amount, "insufficient balance");
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
    use ixc_testing::*;
    use crate::simple_asset::{SimpleAsset, SimpleAssetInit};

    #[test]
    fn test() {
        let mut app = TestApp::default();
        app.register_handler::<SimpleAsset>().unwrap();
        let mut alice = app.new_client_context().unwrap();
        let mut bob = app.new_client_context().unwrap();
        let asset_client = create_account::<SimpleAsset>(&mut alice, SimpleAssetInit{ initial_balance: 100 }).unwrap();
        let alice_balance = asset_client.get_balance(&alice, alice.account_id()).unwrap();
        assert_eq!(alice_balance, 100);
        asset_client.send(&mut alice, 50, bob.account_id()).unwrap();
        let alice_balance = asset_client.get_balance(&alice, alice.account_id()).unwrap();
        assert_eq!(alice_balance, 50);
        let bob_balance = asset_client.get_balance(&bob, bob.account_id()).unwrap();
    }
}

fn main() {}
