#![allow(missing_docs)]
#[interchain_sdk::account_handler]
pub mod counter {
    use interchain_sdk::*;

    pub struct OwnedCounter {
        owner: Item<Address>,
        value: Item<u8>,
    }

    impl OwnedCounter {
        #[on_create]
        fn on_create(&self, ctx: &mut Context) -> Response<()> {
            self.owner.set(ctx, ctx.caller().clone())?;
            Ok(())
        }
    }

    #[publish]
    impl OwnedCounter {
        pub fn get(&self, ctx: &Context) -> Response<u8> {
            self.value.get(ctx)
        }

        pub fn inc(&mut self, ctx: &mut Context) -> Response<()> {
            self.protect(ctx)?;
            let value = self.value.get(ctx)?;
            let new_value = value.checked_add(1).ok_or(())?;
            self.value.set(ctx, value)
        }

        pub fn dec(&mut self, ctx: &mut Context) -> Response<()> {
            self.protect(ctx)?;
            let value = self.value.get(ctx)?;
            let new_value = value.checked_sub(1).ok_or(())?;
            self.value.set(ctx, new_value)
        }

        fn protect(&self, ctx: &Context) -> Response<()> {
            if &self.owner.get(ctx)? != ctx.caller() {
                return Err(());
            }
            Ok(())
        }
    }
}

fn main() {}
