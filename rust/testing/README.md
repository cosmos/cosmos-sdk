**WARNING: This is an API preview! Most code won't work or even type check properly!**

This crate provides a testing framework for Interchain SDK applications.

## Getting Started

This framework works with regular rust `#[test]` functions. To write tests, follow the following steps

1. Add a `use ixc::testing::*` statement (optional, but recommended)
2. Define a `TestApp`
3. Register the handler types we want to test
4. Create account instances for real or mao
5. Perform actions on the accounts and assert the results

Let's define a simple counter with one method `inc` which increments the counter and returns
the new value:
```rust
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
```

Now we can write a test for this counter.
This simple test will create a new counter, increment it
and check the result.

```rust
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
```

## Using Mocks

Mock handlers created by frameworks such as [mockall](https://docs.rs/mockall)
can be used to test the behavior of a handler in isolation.
The [`MockHandler`] type can be used to register boxed mock instances
of one or more [`ixc::handler_api`] traits.
And the [`TestApp::add_mock`] method can be used to add a mock handler to the test app.

See the `examples/` directory in the [`ixc`] crate for more examples on usage.
