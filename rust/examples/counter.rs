#![allow(missing_docs)]
#[interchain_sdk::account_handler(Counter)]
pub mod counter {
    use interchain_sdk::*;

    #[derive(Resources)]
    pub struct Counter {
        value: Item<u64>,
    }

    #[derive(StructCodec)]
    pub struct Test1 {
        pub a: u64,
        pub b: String,
    }

    impl Counter {
        #[on_create]
        pub fn create(ctx: &mut Context) {
        }

        #[publish]
        pub fn get(&self, ctx: &Context) -> Response<u64> {
            // self.value.get(ctx)
            todo!()
        }

        #[publish]
        pub fn inc(&mut self, ctx: &mut Context) -> Response<()> {
            // let value = self.value.get(ctx)?;
            // let new_value = value.checked_add(1).ok_or(())?;
            // self.value.set(ctx, new_value)
            todo!()
        }
    }
}

#[cfg(test)]
mod tests {
    use interchain_core_testing::*;
    use super::counter::*;

    #[test]
    fn test_counter() {
        let mut app = TestApp::default();
        let alice = app.new_client_address();
        let counter_inst = app.add_account::<Counter>(&alice, ()).unwrap();
    }
}

fn main() {}