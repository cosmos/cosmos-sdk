#[interchain_sdk::account_handler]
pub mod counter {
    use interchain_sdk::*;

    pub struct Counter {
        value: Item<u64>,
    }

    #[publish]
    impl Counter {
        pub fn get(&self, ctx: &Context) -> Response<u64> {
            self.value.get(ctx)
        }

        pub fn inc(&mut self, ctx: &mut Context) -> Response<()> {
            let value = self.value.get(ctx)?;
            self.value.set(ctx, value + 1)
        }

        pub fn dec(&mut self, ctx: &mut Context) -> Response<()> {
            let value = self.value.get(ctx)?;
            self.value.set(ctx, value - 1)
        }
    }
}

fn main() {}
