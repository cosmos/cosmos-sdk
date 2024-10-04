#![allow(missing_docs)]
#[ixc::account_handler(OwnerCounter)]
pub mod counter {
    use ixc::*;

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

#[cfg(test)]
mod tests {
    use ixc_testing::*;
    use super::counter::*;

    #[test]
    fn test_counter() {
        let mut app = TestApp::default();
        let mut alice = app.new_client_context();
        let mut bob = app.new_client_context();
        let (counter, _) = app.new_client_context::<OwnedCounter>(&mut alice, ());
        assert_eq!(counter.get(&mut alice).unwrap(), 0);
    }
}

fn main() {}