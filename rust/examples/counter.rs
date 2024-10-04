#![allow(missing_docs)]

use ixc_core::handler::{ClientFactory, Handler, HandlerAPI};
use ixc_core::resource::{Resources, StateObject};
use ixc_core_macros::package_root;
use ixc_message_api::handler::{Allocator, HostBackend, RawHandler};

#[ixc::handler(Counter)]
pub mod counter {
    use ixc::*;
    use ixc_core::resource::{InitializationError, ResourceScope, StateObject};

    // #[derive(Resources)]
    pub struct Counter {
        value: Item<u64>,
    }

    impl Counter {
        #[on_create]
        pub fn create(&self, ctx: &mut Context, init_value: u64) -> Result<()> {
            self.value.set(ctx, init_value)
        }

        #[publish]
        pub fn get(&self, ctx: &Context) -> Result<u64> {
            self.value.get(ctx)
        }

        #[publish]
        pub fn inc(&self, ctx: &mut Context) -> Result<()> {
            let value = self.value.get(ctx)?;
            self.value.set(ctx, value + 1)
        }

        #[publish]
        pub fn add(&self, ctx: &mut Context, value: u64) -> Result<()> {
            let current = self.value.get(ctx)?;
            self.value.set(ctx, current + value)
        }
    }

    unsafe impl Resources for Counter {
        unsafe fn new(scope: &ResourceScope) -> core::result::Result<Self, InitializationError> {
            Ok(Counter {
                value: Item::new(scope.state_scope, 0)?,
            })
        }
    }

    // unsafe impl ixc_core::routes::Router for crate::counter::Counter {
    //     const SORTED_ROUTES: &'static [Route<Self>] =
    //         &sort_routes([
    //             (ON_CREATE_SELECTOR, |counter: &Counter, packet, cb, a| {
    //                 let mut context = Context::new(packet, cb);
    //                 counter.create(&mut context, 42).
    //                     map_err(|e| HandlerError::Custom(0))
    //             }),
    //             (<CounterGetMsg as Message>::SELECTOR, |counter: &Counter, packet, cb, a| {
    //                 let mut context = Context::new(packet, cb);
    //                 let res = counter.get(&mut context).
    //                     map_err(|e| HandlerError::Custom(0))?;
    //                 <u64 as OptionalValue<'_>>::encode_value(&NativeBinaryCodec::default(), &res, packet, a).
    //                     map_err(|e| HandlerError::Custom(0))
    //             }),
    //             (<CounterIncMsg as Message>::SELECTOR, |counter: &Counter, packet, cb, a| {
    //                 let mut context = Context::new(packet, cb);
    //                 counter.inc(&mut context).
    //                     map_err(|e| HandlerError::Custom(0))
    //             }),
    //         ]);
    // }
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
        let alice = app.new_client_account().unwrap();
        let mut alice_ctx = app.client_context_for(alice);
        let counter_client = create_account::<Counter>(&mut alice_ctx, &()).unwrap();
        let cur = counter_client.get(&alice_ctx).unwrap();
        assert_eq!(cur, 0);
        counter_client.inc(&mut alice_ctx).unwrap();
        let cur = counter_client.get(&alice_ctx).unwrap();
        assert_eq!(cur, 1);
        counter_client.add(&mut alice_ctx, 41).unwrap();
        let cur = counter_client.get(&alice_ctx).unwrap();
        assert_eq!(cur, 42);
    }
}

package_root!(counter::Counter);

fn main() {}