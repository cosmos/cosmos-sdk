#![allow(missing_docs)]

use ixc_core::handler::{Handler, HandlerAPI};
use ixc_core::resource::Resources;
use ixc_core::routes::exec_route;
use ixc_core_macros::package_root;
use ixc_message_api::handler::{Allocator, HandlerError, HostBackend, RawHandler};
use ixc_message_api::packet::MessagePacket;
use crate::counter::Counter;

// #[ixc::account_handler(Counter)]
pub mod counter {
    use ixc::*;

    // #[derive(Resources)]
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

impl HandlerAPI for Counter { type ClientFactory = (); }

unsafe impl ixc_core::routes::Router for Counter {
    const SORTED_ROUTES: &'static [ixc_core::routes::Route<Self>] = &[];
}

impl RawHandler for Counter {
    fn handle(&self, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend, allocator: &dyn Allocator) -> Result<(), HandlerError> {
        exec_route(self, message_packet, callbacks, allocator)
    }
}

unsafe impl Resources for Counter {}

impl Handler for Counter {
    const NAME: &'static str = "Counter";
    type Init = u64;
}

#[cfg(test)]
mod tests {
    use ixc_testing::*;
    use super::counter::*;

    #[test]
    fn test_counter() {
        let mut app = TestApp::default();
        let alice = app.new_client_account();
        let alice_ctx = app.client_context_for(alice);
        let counter_inst = app.create_account::<Counter>(alice_ctx, ()).unwrap();
    }
}

// #[cfg(target_arch = "wasm32")]
#[no_mangle]
pub extern fn exec() -> u32 {
    0
}

package_root!(counter::Counter);

fn main() {}