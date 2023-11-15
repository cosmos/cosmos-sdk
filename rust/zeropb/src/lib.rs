#![feature(offset_of)]
#![warn(missing_docs)]
#![no_std]

extern crate core;

mod zerocopy;

pub use zerocopy::ZeroCopy;

mod root;

pub use root::Root;
pub use root::{__zeropb_alloc_page, __zeropb_free_page};

mod error;
mod rel_ptr;

pub use error::Error;

mod str;

pub use str::{Str, StrWriter};

mod bytes;

pub use bytes::{Bytes, BytesWriter};

mod client;
mod r#enum;
mod oneof;
mod ptr;
mod repeated;

pub use repeated::{Repeated, ScalarRepeated, RepeatedWriter, RepeatedIter, ScalarRepeatedWriter};

mod server;
mod status;

pub use status::{Status, Code};

pub use client::{ClientConn, Handler};

#[cfg(test)]
mod tests {
    use core::borrow::Borrow;
    use core::fmt::Write;

    use crate::root::Root;
    use crate::str::Str;
    use crate::zerocopy::ZeroCopy;

    #[repr(C)]
    struct TestStruct {
        s: Str,
        x: u32,
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

    struct MsgSend {
        from: Str,
        to: Str,
        denom: Str,
        amount: u64,
    }

    unsafe impl ZeroCopy for MsgSend {}

    struct MsgSendResponse {}

    unsafe impl ZeroCopy for MsgSendResponse {}

    enum Error {}

    trait MsgServer {
        fn send(&mut self, msg: &MsgSend, response: &mut MsgSendResponse) -> Result<(), Error>;
    }
}
