This crate provides a testing framework for Interchain SDK applications.

## Writing Tests

1. Start by writing tests as you would normally do in Rust.
2. Define a `TestApp`
3. Add real or mock modules & accounts to the `TestApp`
4. Perform actions on the modules & accounts and assert the results

Let's define a simple counter with two methods `get` and `inc`:
```rust
#[interchain_sdk::account_handler(Counter)]
pub mod counter {
    use interchain_sdk::*;

    #[derive(Resources)]
    pub struct Counter {
        value: Item<u64>,
    }

    impl OnCreate for Counter {
        type InitMessage = ();

        fn on_create(&self, ctx: &mut std::task::Context, init: &Self::InitMessage) -> Response<()> {
            Ok(())
        }
    }

    #[publish]
    impl Counter {
        pub fn get(&self, ctx: &Context) -> Response<u64> {
            self.value.get(ctx)
        }

        pub fn inc(&mut self, ctx: &mut Context) -> Response<()> {
            let value = self.value.get(ctx)?;
            let new_value = value.checked_add(1).ok_or(())?;
            self.value.set(ctx, new_value)
        }
    }
}
```
