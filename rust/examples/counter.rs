#![allow(missing_docs)]

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

const HANDLER_0: counter::Counter = unsafe { <counter::Counter as ::ixc::core::resource::Resources>::new(&<::ixc::core::resource::ResourceScope as core::default::Default>::default()).unwrap() };

const HANDLER_DESCRIPTOR_0: &[RawHandlerDescriptor] = &[RawHandlerDescriptor {
    name: b"Counter\0".as_ptr(),
    name_len: 7,
}];

type Callback = fn(ctx: u64, packet: *mut u8, packet_len: u32) -> u32;

#[no_mangle]
pub extern "C" fn ixc_handle(handler_idx: u32, packet: *mut u8, packet_len: u32, ctx: u64, callback: Callback) -> u32 {
    todo!()
}

#[no_mangle]
pub extern "C" fn ixc_num_handlers() -> u32 {
    todo!()
}

#[no_mangle]
pub extern "C" fn ixc_describe_handler(handler_idx: u32) -> *const RawHandlerDescriptor {
    todo!()
}

#[repr(C)]
struct RawHandlerDescriptor {
    name: *const u8,
    name_len: u32,
}

#[no_mangle]
pub extern "C" fn ixc_alloc(size: u32, align: u32) -> *mut u8 {
    todo!()
}

#[no_mangle]
pub extern "C" fn ixc_free(ptr: *mut u8, size: u32) {
    todo!()
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
        let counter_client = create_account::<Counter>(&mut alice_ctx, CounterCreate{}).unwrap();
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

fn main() {}