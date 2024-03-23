#![no_std]

use alloc::format;
use alloc::vec::Vec;
use core::ptr::slice_from_raw_parts;
use core::fmt::Write;

pub mod test1 {
    include!(concat!(env!("OUT_DIR"), "/test1.rs"));
}

pub mod cosmos {
    pub mod base {
        pub mod v1beta1 {
            include!(concat!(env!("OUT_DIR"), "/cosmos.base.v1beta1.rs"));
        }
    }

    pub mod bank {
        pub mod v1beta1 {
            include!(concat!(env!("OUT_DIR"), "/cosmos.bank.v1beta1.rs"));
        }
    }
}

extern crate alloc;

#[cfg(target_arch = "wasm32")]
use lol_alloc::{FreeListAllocator, LockedAllocator};

#[cfg(target_arch = "wasm32")]
#[global_allocator]
static ALLOCATOR: LockedAllocator<FreeListAllocator> = LockedAllocator::new(FreeListAllocator::new());

#[cfg(not(target_arch = "wasm32"))]
extern crate std;

#[cfg(target_arch = "wasm32")]
#[panic_handler]
fn panic(_info: &core::panic::PanicInfo) -> ! {
    loop {}
}

use prost::Message;

#[cfg(target_arch = "wasm32")]
#[no_mangle]
pub extern fn exec(input: *const u8, len: i32) -> i64 {
    unsafe {
        let (ptr, len) = do_exec(input, len as usize);
        let res: i64 = ptr as i64 | ((len as i64) << 32);
        res
    }
}

#[cfg(not(target_arch = "wasm32"))]
#[no_mangle]
pub extern fn exec(input: *const u8, len: usize, out_len: *mut usize) -> *const u8 {
    unsafe {
        let (ptr, len) = do_exec(input, len as usize);
        *out_len = len;
        ptr
    }
}

unsafe fn do_exec(input: *const u8, len: usize) -> (*const u8, usize) {
    let bz = slice_from_raw_parts(input, len );
    let msg = test1::Greet::decode(&*bz).unwrap();
    let res = test1::GreetResponse{
        message: format!("Hello, {}! You entered {}", msg.name, msg.value),
    };
    let buf = res.encode_to_vec();
    let len = buf.len();
    let ptr = buf.as_ptr();
    core::mem::forget(buf);
    (ptr, len)
}

#[cfg(target_arch = "wasm32")]
#[no_mangle]
pub extern fn exec_msg_send(input: *const u8, len: i32) -> i64 {
    unsafe {
        let (ptr, len) = do_exec_msg_send(input, len as usize);
        let res: i64 = ptr as i64 | ((len as i64) << 32);
        res
    }
}

#[cfg(not(target_arch = "wasm32"))]
#[no_mangle]
pub extern fn exec_msg_send(input: *const u8, len: usize, out_len: *mut usize) -> *const u8 {
    unsafe {
        let (ptr, len) = do_exec_msg_send(input, len as usize);
        *out_len = len;
        ptr
    }
}

unsafe fn do_exec_msg_send(input: *const u8, len: usize) -> (*const u8, usize) {
    let bz = slice_from_raw_parts(input, len );
    let msg_send = cosmos::bank::v1beta1::MsgSend::decode(&*bz).unwrap();
    let mut msg = format!("{} sent to {}:", &msg_send.from_address, &msg_send.to_address);
    for coin in msg_send.amount.into_iter() {
        let coin_msg = format!("  {} {}", &coin.amount, &coin.denom);
        write!(msg, "{}", coin_msg).unwrap();
    }
    let res = test1::GreetResponse{
        message: msg,
    };
    let buf = res.encode_to_vec();
    let len = buf.len();
    let ptr = buf.as_ptr();
    core::mem::forget(buf);
    (ptr, len)
}

#[no_mangle]
pub extern fn __alloc(size: usize) -> *mut u8 {
    unsafe {
        let mut buf = Vec::with_capacity(size);
        let ptr = buf.as_mut_ptr();
        core::mem::forget(buf);
        ptr
    }
}

#[no_mangle]
pub extern fn __free(ptr: *mut u8, size: usize) {
    unsafe {
        let _buf = Vec::from_raw_parts(ptr, 0, size);
    }
}
