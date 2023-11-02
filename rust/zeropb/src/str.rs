use core::marker::PhantomData;
use core::{
    borrow::Borrow,
    fmt::Write,
    str::from_utf8_unchecked,
};

use crate::bytes::{Bytes, BytesWriter};
use crate::error::Error;
use crate::zerocopy::ZeroCopy;

#[repr(C)]
pub struct Str {
    pub(crate) ptr: Bytes,
    _phantom: PhantomData<str>,
}

unsafe impl ZeroCopy for Str {}

impl Str {
    pub fn set(&mut self, content: &str) -> Result<(), Error> {
        self.ptr.set(content.as_bytes())
    }

    pub fn new_writer(&mut self) -> Result<StrWriter, Error> {
        self.ptr.new_writer().map(|bz| StrWriter { bz })
    }
}

impl<'a> Borrow<str> for Str {
    fn borrow(&self) -> &str {
        unsafe {
            from_utf8_unchecked(self.ptr.borrow())
        }
    }
}


pub struct StrWriter<'a> {
    bz: BytesWriter<'a>,
}

impl <'a> Write for StrWriter<'a> {
    fn write_str(&mut self, s: &str) -> core::fmt::Result {
        self.bz.write(s.as_bytes()).map_err(|_| core::fmt::Error)
    }
}

