#![allow(missing_docs)]

use ixc_core::account_api::ON_CREATE_SELECTOR;
use ixc_core::Context;
use ixc_core::handler::{ClientFactory, Handler, HandlerAPI};
use ixc_core::resource::{InitializationError, ResourceScope, Resources, StateObject};
use ixc_core::routes::{exec_route, sort_routes, Route};
use ixc_core_macros::{message_selector, package_root};
use ixc_message_api::AccountID;
use ixc_message_api::handler::{Allocator, HandlerError, HostBackend, RawHandler};
use ixc_message_api::packet::MessagePacket;
use state_objects::Item;
use crate::counter::Counter;

#[ixc::handler(Counter)]
pub mod counter {
    use ixc::*;
    use ixc_core::resource::{InitializationError, ResourceScope, StateObject};

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

    unsafe impl Resources for Counter {
        unsafe fn new(scope: &ResourceScope) -> core::result::Result<Self, InitializationError> {
            Ok(Counter {
                value: Item::new(scope.state_scope, 0)?,
            })
        }
    }
}

unsafe impl ixc_core::routes::Router for Counter {
    const SORTED_ROUTES: &'static [Route<Self>] =
        &sort_routes([
            (ON_CREATE_SELECTOR, |counter: &Counter, packet, cb, a| {
                let mut context = Context::new(packet, cb);
                counter.create(&mut context, 0);
                todo!()
            }),
            // (message_selector!("get"), |counter, ctx| counter.get(ctx)),
            // (message_selector!("inc"), |counter, ctx| counter.inc(ctx)),
        ]);
}

// impl RawHandler for Counter {
//     fn handle(&self, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend, allocator: &dyn Allocator) -> Result<(), HandlerError> {
//         exec_route(self, message_packet, callbacks, allocator)
//     }
// }

#[cfg(test)]
mod tests {
    use ixc_core::account_api::create_account;
    use ixc_core::handler::Handler;
    use ixc_testing::*;
    use super::counter::*;

    #[test]
    fn test_counter() {
        let mut app = TestApp::default();
        app.register_handler::<Counter>().unwrap();
        let alice = app.new_client_account().unwrap();
        let mut alice_ctx = app.client_context_for(alice);
        let counter_client = create_account::<Counter>(&mut alice_ctx, &()).unwrap();
    }
}

// #[cfg(target_arch = "wasm32")]
#[no_mangle]
pub extern fn exec() -> u32 {
    0
}

package_root!(counter::Counter);

fn main() {}