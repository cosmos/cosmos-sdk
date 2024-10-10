#![allow(missing_docs)]

#[ixc::handler(Counter)]
pub mod counter {
    use ixc::*;

    #[derive(Resources)]
    pub struct Counter {
        #[state]
        value: Accumulator,
    }

    impl Counter {
        #[on_create]
        pub fn create(&self, ctx: &mut Context) -> Result<()> {
            Ok(())
        }

        #[publish]
        pub fn get(&self, ctx: &Context) -> Result<u128> {
            let res = self.value.get(ctx)?;
            Ok(res)
        }

        #[publish]
        pub fn inc(&self, ctx: &mut Context) -> Result<u128> {
            let value = self.value.add(ctx, 1)?;
            Ok(value)
        }

        #[publish]
        pub fn dec(&self, ctx: &mut Context) -> Result<u128> {
            let value = self.value.safe_sub(ctx, 1)?;
            Ok(value)
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
        let counter_client = create_account(&mut alice_ctx, CounterCreate{}).unwrap();
        let cur = counter_client.get(&alice_ctx).unwrap();
        assert_eq!(cur, 0);
        let cur = counter_client.inc(&mut alice_ctx).unwrap();
        assert_eq!(cur, 1);
        let cur = counter_client.inc(&mut alice_ctx).unwrap();
        assert_eq!(cur, 2);
        let cur = counter_client.dec(&mut alice_ctx).unwrap();
        assert_eq!(cur, 1);
        let cur = counter_client.dec(&mut alice_ctx).unwrap();
        assert_eq!(cur, 0);
        let res = counter_client.dec(&mut alice_ctx);
        assert!(res.is_err());
    }
}

ixc::package_root!(counter::Counter);

fn main() {}