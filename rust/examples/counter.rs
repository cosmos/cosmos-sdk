#![allow(missing_docs)]

use crate::counter::{Counter, CounterClient};
use ixc_core::account_api::ON_CREATE_SELECTOR;
use ixc_core::handler::{ClientFactory, Handler, HandlerAPI};
use ixc_core::resource::{Resources, StateObject};
use ixc_core::routes::{sort_routes, Route};
use ixc_core::Context;
use ixc_core_macros::package_root;
use ixc_message_api::handler::{Allocator, HandlerError, HostBackend, RawHandler};

#[ixc::handler(Counter)]
pub mod counter {
    use ixc::*;
    use ixc_core::account_api::ON_CREATE_SELECTOR;
    use ixc_core::low_level::create_packet;
    use ixc_core::message::Message;
    use ixc_core::resource::{InitializationError, ResourceScope, StateObject};
    use ixc_core::routes::{sort_routes, Route};
    use ixc_message_api::handler::{Allocator, HandlerError, HandlerErrorCode, HostBackend, RawHandler};
    use ixc_message_api::header::MessageSelector;
    use ixc_schema::binary::NativeBinaryCodec;
    use ixc_schema::codec::decode_value;
    use ixc_schema::value::OptionalValue;

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
    }

    unsafe impl Resources for Counter {
        unsafe fn new(scope: &ResourceScope) -> core::result::Result<Self, InitializationError> {
            Ok(Counter {
                value: Item::new(scope.state_scope, 0)?,
            })
        }
    }

    impl CounterClient {
        // pub fn get(&self, ctx: &Context) -> ixc_core::Result<u64> {
        //     let mut packet = create_packet(ctx, self.0, GET_SELECTOR)?;
        //     unsafe {
        //         ctx.host_backend().invoke(&mut packet, ctx.memory_manager())
        //             .map_err(|e| ())?;
        //             // .map_err(|e| fmt_error!("unknown error: {:?}", e))?;
        //         let cdc = NativeBinaryCodec::default();
        //         let value = <u64 as OptionalValue<'_>>::decode_value(&cdc, &packet, ctx.memory_manager())
        //             .map_err(|e| ())?;
        //             // .map_err(|e| fmt_error!("decoding error: {:?}", e))?;
        //         Ok(value)
        //     }
        // }
        //
        // pub fn inc(&self, ctx: &mut Context) -> Result<()> {
        //     let mut packet = create_packet(ctx, self.0, INC_SELECTOR)?;
        //     unsafe {
        //         ctx.host_backend().invoke(&mut packet, ctx.memory_manager())
        //             .map_err(|e| ())
        //     }
        // }
    }

    unsafe impl ixc_core::routes::Router for crate::counter::Counter {
        const SORTED_ROUTES: &'static [Route<Self>] =
            &sort_routes([
                (ON_CREATE_SELECTOR, |counter: &Counter, packet, cb, a| {
                    let mut context = Context::new(packet, cb);
                    counter.create(&mut context, 42).
                        map_err(|e| HandlerError::Custom(0))
                }),
                (<CounterGetMsg as Message>::SELECTOR, |counter: &Counter, packet, cb, a| {
                    let mut context = Context::new(packet, cb);
                    let res = counter.get(&mut context).
                        map_err(|e| HandlerError::Custom(0))?;
                    <u64 as OptionalValue<'_>>::encode_value(&NativeBinaryCodec::default(), &res, packet, a).
                        map_err(|e| HandlerError::Custom(0))
                }),
                (<CounterIncMsg as Message>::SELECTOR, |counter: &Counter, packet, cb, a| {
                    let mut context = Context::new(packet, cb);
                    counter.inc(&mut context).
                        map_err(|e| HandlerError::Custom(0))
                }),
            ]);
    }

    // const INC_SELECTOR = message_selector!("inc");
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
        assert_eq!(cur, 42);
        counter_client.inc(&mut alice_ctx).unwrap();
        let cur = counter_client.get(&alice_ctx).unwrap();
        assert_eq!(cur, 43);
    }
}

package_root!(counter::Counter);

fn main() {}