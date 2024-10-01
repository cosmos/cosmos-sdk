#![allow(missing_docs)]

use ixc_core_macros::package_root;
use crate::counter::Counter;

// #[ixc::account_handler(Counter)]
pub mod counter {
    use ixc::*;

    #[derive(Resources)]
    pub struct Counter {
        value: Item<u64>,
    }

    impl Counter {
        // #[on_create]
        pub fn create(&self, ctx: &mut Context, init_value: u64) -> Result<()> {
            self.value.set(ctx, init_value)
        }

        // #[publish]
        pub fn get(&self, ctx: &Context) -> Result<u64> {
            self.value.get(ctx)
        }

        // #[publish]
        pub fn inc(&mut self, ctx: &mut Context) -> Result<()> {
            let value = self.value.get(ctx)?;
            let new_value = value.checked_add(1).ok_or(
                // fmt_error!("overflow when incrementing counter")
                todo!()
            )?;
            self.value.set(ctx, new_value)
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
        let alice = app.new_client_address();
        let counter_inst = app.add_account::<Counter>(&alice, ()).unwrap();
    }
}

// #[cfg(target_arch = "wasm32")]
#[no_mangle]
pub extern fn exec() -> u32 {
    0
}

package_root!(counter::Counter);

fn main() {}