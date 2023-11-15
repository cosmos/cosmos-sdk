#![no_std]

use zeropb::{__zeropb_alloc_page, __zeropb_free_page, Root};
use core::fmt::Write;

pub mod test1 {
    include!(concat!(env!("OUT_DIR"), "/out.rs"));
}

#[no_mangle]
pub extern fn exec(input: *mut u8, len: i32) -> i64 {
    unsafe {
        let req = Root::<test1::Greet>::unsafe_wrap(input);
        let mut res = Root::<test1::GreetResponse>::new();
        write!(res.message.new_writer().unwrap(), "Hello, {}! You entered {}", &req.name, req.value).unwrap();
        res.unsafe_unwrap() as i64
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
