#![allow(missing_docs)]

#[ixc::handler(Counter)]
pub mod counter {
    use ixc::*;
    use ixc_core::resource::{InitializationError, ResourceScope, StateObjectResource};

    #[derive(Resources)]
    pub struct Counter {
        #[state]
        value: Item<u64>,
    }

    impl Counter {
        #[on_create]
        pub fn create(&self, ctx: &mut Context, init_value: u64) -> Result<()> {
            self.value.set(ctx, init_value)
        }

        #[publish]
        pub fn get(&self, ctx: &Context) -> Result<u64> {
            self.value.get(ctx)
        }

        #[publish]
        pub fn inc(&self, ctx: &mut Context) -> Result<()> {
            let value = self.value.get(ctx)?;
            self.value.set(ctx, value + 1)
        }

        #[publish]
        pub fn add(&self, ctx: &mut Context, value: u64) -> Result<()> {
            let current = self.value.get(ctx)?;
            self.value.set(ctx, current + value)
        }
    }
}


#[cfg(test)]
mod tests {
    use super::counter::*;
    use ixc_core::account_api::create_account;
    use ixc_testing::*;

    #[test]
    fn test_counter() {
        let mut app = TestApp::default();
        app.register_handler::<Counter>().unwrap();
        let mut alice_ctx = app.new_client_context().unwrap();
        let counter_client = create_account(&mut alice_ctx, CounterCreate{init_value: 41}).unwrap();
        let cur = counter_client.get(&alice_ctx).unwrap();
        assert_eq!(cur, 41);
        counter_client.inc(&mut alice_ctx).unwrap();
        let cur = counter_client.get(&alice_ctx).unwrap();
        assert_eq!(cur, 42);
        counter_client.add(&mut alice_ctx, 12).unwrap();
        let cur = counter_client.get(&alice_ctx).unwrap();
        assert_eq!(cur, 54);
    }
}

ixc::package_root!(counter::Counter);

fn main() {}