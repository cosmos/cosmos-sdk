use core::fmt::Debug;
use crate::{Code, Root, Str};

pub struct Error {
    pub(crate) code: Code,
    pub(crate) msg: Root<Str>,
}

impl Debug for Error {
    fn fmt(&self, f: &mut core::fmt::Formatter<'_>) -> core::fmt::Result {
        f.debug_struct("Error")
            .field("code", &self.code)
            .finish()
    }
}

impl PartialEq for Error {
    fn eq(&self, other: &Self) -> bool {
        self.code == other.code
    }
}

// #[derive(Error, Debug)]
// pub enum Error {
//     #[error("out of memory")]
//     OutOfMemory,
//
//     #[error("out of bounds")]
//     OutOfBounds,
//
//     #[error("invalid state")]
//     InvalidState,
//
//     #[error("invalid buffer")]
//     InvalidBuffer,
// }
