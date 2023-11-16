#![no_std]

use zeropb::{__zeropb_alloc_page, __zeropb_free_page, Root};
use core::fmt::Write;
use core::mem::forget;

pub mod test1 {
    include!(concat!(env!("OUT_DIR"), "/out.rs"));
}

#[cfg(target_arch = "wasm32")]
#[no_mangle]
pub extern fn exec(input: *mut u8, len: i32) -> i64 {
    unsafe {
        let req = Root::<test1::Greet>::unsafe_wrap(input);
        let mut res = Root::<test1::GreetResponse>::new();
        write!(res.message.new_writer().unwrap(), "Hello, {}! You entered {}", &req.name, req.value).unwrap();
        res.unsafe_unwrap() as i64
    }
}

extern crate std;
use std::println;

#[cfg(not(target_arch = "wasm32"))]
#[no_mangle]
pub extern fn exec(input: *mut u8, len: usize, out_len: *mut usize) -> *const u8 {
    unsafe {
        let req = Root::<test1::Greet>::unsafe_wrap(input);
        let mut res = Root::<test1::GreetResponse>::new();
        write!(res.message.new_writer().unwrap(), "Hello, {}! You entered {}", &req.name, req.value).unwrap();
        let out = res.as_slice();
        *out_len = out.len();
        let out = out.as_ptr();
        forget(req);
        forget(res);
        out
    }
}

#[no_mangle]
pub extern fn __alloc(size: usize) -> *mut u8 {
    __zeropb_alloc_page()
}

#[no_mangle]
pub extern fn __free(ptr: *mut u8, size: usize) {
    __zeropb_free_page(ptr)
}


#[cfg(test)]
mod tests {
    use zeropb::Root;
    extern crate std;
    use std::println;
    use crate::test1;

    #[test]
    fn test() {
        let mut req = Root::<test1::Greet>::new();
        req.name.set("Benchmarker").unwrap();
        req.value = 51;
        println!("req: {:#02x?}", req.as_slice());
    }
}