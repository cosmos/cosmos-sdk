#![no_std]

pub mod nft {
    include!(concat!(env!("OUT_DIR"), "/out.rs"));
}

pub use cosmossdk_core;

#[no_mangle]
pub extern fn add(left: usize, right: usize) -> usize {
    left + right
}

#[no_mangle]
pub extern fn foo(a: *const u8) -> *const u8 {
    cosmossdk_core::test1();
    a
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn it_works() {
        let result = add(2, 2);
        assert_eq!(result, 4);
    }
}