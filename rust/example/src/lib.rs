//! This crate exists for the sole purpose of making sure that macros work correctly in a single
//! dependency scenario.
#![allow(missing_docs)]

#[ixc::handler(Counter)]
pub mod counter {
    use ixc::*;

    #[derive(Resources)]
    pub struct Counter {
        #[state]
        value: Item<u64>,
    }

    #[publish]
    impl Counter {
        #[on_create]
        pub fn create(&self, ctx: &mut Context) -> Result<()> { Ok(()) }

        pub fn inc(&self, ctx: &mut Context) -> Result<u64> {
            let value = self.value.get(ctx)?;
            let new_value = value.checked_add(1).ok_or(
                error!("overflow when incrementing counter")
            )?;
            self.value.set(ctx, new_value)?;
            Ok(new_value)
        }
    }
}



#[cfg(test)]
mod tests {
    use super::counter::*;
    use ixc_testing::*;

    #[test]
    fn test_counter() {
        // create the test app
        let mut app = TestApp::default();
        // register the Counter handler type
        app.register_handler::<Counter>().unwrap();
        // create a new client context for a random user Alice
        let mut alice_ctx = app.new_client_context().unwrap();
        // Alice creates a new counter account
        let counter_client = create_account::<Counter>(&mut alice_ctx, CounterCreate {}).unwrap();
        // Alice increments the counter
        let value = counter_client.inc(&mut alice_ctx).unwrap();
        assert_eq!(value, 1);
        // Alice increments the counter again
        let value = counter_client.inc(&mut alice_ctx).unwrap();
        assert_eq!(value, 2);
    }
}