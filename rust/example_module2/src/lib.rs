#![no_std]

use alloc::format;
use alloc::vec::Vec;
use core::ptr::slice_from_raw_parts;

pub mod test1 {
    include!(concat!(env!("OUT_DIR"), "/test1.rs"));
}

extern crate alloc;

#[cfg(target_arch = "wasm32")]
use lol_alloc::{FreeListAllocator, LockedAllocator};

#[cfg(target_arch = "wasm32")]
#[global_allocator]
static ALLOCATOR: LockedAllocator<FreeListAllocator> = LockedAllocator::new(FreeListAllocator::new());

#[panic_handler]
fn panic(_info: &core::panic::PanicInfo) -> ! {
    loop {}
}

use prost::Message;

#[no_mangle]
pub extern fn exec(input: *const u8, len: i32) -> i64 {
    unsafe {
        let bz = slice_from_raw_parts(input, len as usize);
        let msg = test1::Greet::decode(&*bz).unwrap();
        let res = test1::GreetResponse{
            message: format!("Hello, {}! You entered {}", msg.name, msg.value),
        };
        let buf = res.encode_to_vec();
        let len = buf.len();
        let ptr = buf.as_ptr();
        core::mem::forget(buf);
        let res: i64 = ptr as i64 | ((len as i64) << 32);
        res
    }
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
