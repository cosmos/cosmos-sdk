#![allow(missing_docs)]

use std::ptr::NonNull;
use ixc_message_api::code::ErrorCode;
use ixc_message_api::handler::{Allocator, RawHandler};
use ixc_message_api::header::MessageHeader;
use ixc_message_api::packet::MessagePacket;

#[ixc::handler(Counter)]
pub mod counter {
    use ixc::*;

    #[derive(Resources)]
    pub struct Counter {
        #[state]
        value: Accumulator,
    }

    impl Counter {
        #[on_create]
        pub fn create(&self, ctx: &mut Context) -> Result<()> {
            Ok(())
        }

        #[publish]
        pub fn get(&self, ctx: &Context) -> Result<u128> {
            let res = self.value.get(ctx)?;
            Ok(res)
        }

        #[publish]
        pub fn inc(&self, ctx: &mut Context) -> Result<u128> {
            let value = self.value.add(ctx, 1)?;
            Ok(value)
        }

        #[publish]
        pub fn dec(&self, ctx: &mut Context) -> Result<u128> {
            let value = self.value.safe_sub(ctx, 1)?;
            Ok(value)
        }
    }
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
        let mut alice_ctx = app.new_client_context().unwrap();
        let counter_client = create_account(&mut alice_ctx, CounterCreate{}).unwrap();
        let cur = counter_client.get(&alice_ctx).unwrap();
        assert_eq!(cur, 0);
        let cur = counter_client.inc(&mut alice_ctx).unwrap();
        assert_eq!(cur, 1);
        let cur = counter_client.inc(&mut alice_ctx).unwrap();
        assert_eq!(cur, 2);
        let cur = counter_client.dec(&mut alice_ctx).unwrap();
        assert_eq!(cur, 1);
        let cur = counter_client.dec(&mut alice_ctx).unwrap();
        assert_eq!(cur, 0);
        let res = counter_client.dec(&mut alice_ctx);
        assert!(res.is_err());
    }
}

ixc::package_root!(counter::Counter);

#[cfg(target_arch = "wasm32")]
#[no_mangle]
unsafe extern "C" fn handle(packet: *mut u8, len: usize) -> u32 {
    let scope = ixc_core::resource::ResourceScope::default();
    let handler = <crate::counter::Counter as ixc_core::resource::Resources>::new(&scope).unwrap();
    let mut packet = ::ixc_message_api::packet::MessagePacket::new(NonNull::new_unchecked(packet as *mut MessageHeader), len);
    struct Callbacks;
    impl ::ixc_message_api::handler::HostBackend for Callbacks {
        fn invoke(&self, message_packet: &mut MessagePacket, allocator: &dyn Allocator) -> Result<(), ErrorCode> {
            let (ptr, size) = unsafe { message_packet.raw_parts() };
            let code: u32 = unsafe { invoke(ptr.as_ptr() as *const u8, size) };
            if code != 0 {
                Err(ErrorCode::from(code as u16))
            } else {
                Ok(())
            }
        }
    }
    let res = handler.handle(&mut packet, &Callbacks, &allocator_api2::alloc::Global);
    match res {
        Ok(()) => 0,
        Err(code) => {
            let c: u16 = code.into();
            c as u32
        }
    }
}

#[cfg(target_arch = "wasm32")]
#[no_mangle]
unsafe extern "C" fn alloc(size: usize, align: usize) -> *mut u8 {
    std::alloc::alloc(std::alloc::Layout::from_size_align(size, align).unwrap())
}

#[cfg(target_arch = "wasm32")]
#[no_mangle]
unsafe extern "C" fn free(ptr: *mut u8, size: usize) {
    std::alloc::dealloc(ptr, std::alloc::Layout::from_size_align(size, 1).unwrap())
}

#[cfg(target_arch = "wasm32")]
#[link(wasm_import_module = "ixc")]
extern "C" {
    fn invoke(packet: *const u8, len: usize) -> u32;
}

fn main() {}