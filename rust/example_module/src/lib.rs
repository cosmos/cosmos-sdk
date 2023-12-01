#![no_std]

use zeropb::root::{alloc_page, free_page, Root};
use core::fmt::Write;
use core::mem::forget;
use core::iter::IntoIterator;
use crate::cosmos::bank::v1beta1::MsgSend;
use crate::test1::GreetResponse;

pub mod test1 {
    include!(concat!(env!("OUT_DIR"), "/test1/test.rs"));
}

pub mod cosmos {
    pub mod base {
        pub mod v1beta1 {
            include!(concat!(env!("OUT_DIR"), "/cosmos/base/v1beta1/coin.rs"));
        }
    }

    pub mod bank {
        pub mod v1beta1 {
            include!(concat!(env!("OUT_DIR"), "/cosmos/bank/v1beta1/bank.rs"));
            include!(concat!(env!("OUT_DIR"), "/cosmos/bank/v1beta1/tx.rs"));
        }
    }
}

#[cfg(target_arch = "wasm32")]
#[no_mangle]
pub extern fn exec(input: *mut u8, len: i32) -> i64 {
    unsafe {
        let req = Root::<test1::Greet>::unsafe_wrap(input);
        let mut res = Root::<test1::GreetResponse>::new();
        write!(res.message.new_writer().unwrap(), "Hello, {}! You entered {}", &req.name, req.value).unwrap();
        let res_ptr = res.unsafe_unwrap() as i64;
        forget(req);
        forget(res);
        res_ptr
    }
}

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

#[cfg(target_arch = "wasm32")]
#[no_mangle]
pub unsafe extern fn exec_msg_send(input: *mut u8, len: i32) -> i64 {
    let res = do_exec_msg_send(input);
    let res_ptr = res.unsafe_unwrap() as i64;
    forget(res);
    res_ptr
}

#[cfg(not(target_arch = "wasm32"))]
#[no_mangle]
pub unsafe extern fn exec_msg_send(input: *mut u8, len: usize, out_len: *mut usize) -> *const u8 {
    let res = do_exec_msg_send(input);
    let out = res.as_slice();
    *out_len = out.len();
    let out = out.as_ptr();
    forget(res);
    out
}

unsafe fn do_exec_msg_send(input: *mut u8) -> Root<GreetResponse> {
    let req = Root::<cosmos::bank::v1beta1::MsgSend>::unsafe_wrap(input);
    let mut res = Root::<test1::GreetResponse>::new();
    write!(res.message.new_writer().unwrap(), "{} sent to {}:", &req.from_address, &req.to_address).unwrap();
    for coin in req.amount.into_iter() {
        write!(res.message.new_writer().unwrap(), "  {} {}", &coin.amount, &coin.denom).unwrap();
    }
    forget(req);
    res
}

#[no_mangle]
pub unsafe extern fn __alloc(size: usize) -> *mut u8 {
    alloc_page()
}

#[no_mangle]
pub unsafe extern fn __free(ptr: *mut u8, size: usize) {
    free_page(ptr)
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

    #[test]
    fn test_msg_send() {
        let mut req = Root::<crate::cosmos::bank::v1beta1::MsgSend>::new();
        req.from_address.set("cosmos1huydeevpz37sd9snkgul6070mstupukw00xkw9").unwrap();
        req.to_address.set("cosmos1xy4yqngt0nlkdcenxymg8tenrghmek4nmqm28k").unwrap();
        let mut coins = req.amount.start_write().unwrap();
        let mut coin = coins.append().unwrap();
        coin.denom.set("uatom").unwrap();
        coin.amount.set("1234567").unwrap();
        let mut coin = coins.append().unwrap();
        coin.denom.set("foo").unwrap();
        coin.amount.set("7654321").unwrap();
        println!("req: {:#02x?}", req.as_slice());
    }
}