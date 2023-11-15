#![no_std]

pub mod nft {
    include!(concat!(env!("OUT_DIR"), "/out.rs"));
}

// pub use cosmossdk_core;

#[no_mangle]
pub extern "C" fn add(left: usize, right: usize) -> usize {
    left + right
}

#[no_mangle]
pub extern "C" fn foo(a: *const u8) -> &'static str {
    // cosmossdk_core::test1();
    "abc"
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
