#![feature(offset_of)]
#![warn(missing_docs)]

extern crate alloc;
extern crate core;

mod zerocopy;
mod root;
mod util;
mod error;
mod str;
mod bytes;
mod repeated;
mod oneof;
mod r#enum;
mod ptr;


#[cfg(test)]
mod tests {
    use core::fmt::Write;
    use std::borrow::Borrow;

    use crate::root::Root;
    use crate::str::Str;
    use crate::zerocopy::ZeroCopy;

    #[repr(C)]
    struct TestStruct {
        s: Str,
    }

    unsafe impl ZeroCopy for TestStruct {}

    #[test]
    fn test_str_set() {
        let mut r = Root::<TestStruct>::new();
        r.s.set("hello").unwrap();
        assert_eq!(<Str as Borrow<str>>::borrow(&r.s), "hello");
    }

    #[test]
    fn test_str_writer() {
        let mut r = Root::<TestStruct>::new();
        let mut w = r.s.new_writer().unwrap();
        w.write_str("hello").unwrap();
        w.write_str(" world").unwrap();
        assert_eq!(<Str as Borrow<str>>::borrow(&r.s), "hello world");
    }
}
