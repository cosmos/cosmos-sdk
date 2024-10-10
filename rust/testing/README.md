**WARNING: This is an API preview! Most code won't work or even type check properly!**

This crate provides a testing framework for Interchain SDK applications.

## Getting Started

This framework works with regular rust `#[test]` functions. To write tests, follow the following steps

1. Add a `use ixc::testing::*` statement (optional, but recommended)
2. Define a `TestApp`
3. Register the handler types we want to test
4. Create account instances for real or mock handlers
5. Perform actions on the accounts and assert the results

Let's define a simple counter with one method `inc` which increments the counter and returns
the new value:
```rust
#[ixc::handler(Counter)]
pub mod counter {
    use ixc::*;

    #[derive(Resources)]
    pub struct Counter {
        value: Item<u64>,
    }

    #[publish]
    impl Counter {
        #[on_create]
        pub fn create(&mut self, ctx: &mut Context) -> Result<()> { Ok(()) }
        
        pub fn inc(&mut self, ctx: &mut Context) -> Result<u64> {
            let value = self.value.get(ctx)?;
            let new_value = value.checked_add(1).ok_or(
                bail!("overflow when incrementing counter")
            )?;
            self.value.set(ctx, new_value)?;
            Ok(new_value)
        }
    }
}
```

Now we can write a test for this counter:
```rust
#[cfg(test)]
mod tests {
    use ixc::testing::*;
    use crate::counter::*;
    
    #[test]
    fn test() {
        let mut app = TestApp::new();
    }
}
```
