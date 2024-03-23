extern crate core;
use core::ptr::null;
use crate::Code;
use crate::raw::RawString;
use core::convert::From;
use core::default::Default;

pub struct Error {
    pub code: Code,
    pub message: RawString,
}

impl From<Code> for Error {
    fn from(value: Code) -> Self {
        Error {
            code: value,
            message: RawString::default(),
        }
    }
}