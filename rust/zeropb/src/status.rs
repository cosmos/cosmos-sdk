use crate::r#enum::{Enum, ZeroCopyEnum};
use crate::str::Str;
use crate::zerocopy::ZeroCopy;
use num_enum::{IntoPrimitive, TryFromPrimitive};

#[repr(u8)]
#[derive(Clone, Copy, IntoPrimitive, TryFromPrimitive, Eq, PartialEq, Debug)]
enum Code {
    Ok = 0,
    Cancelled = 1,
    Unknown = 2,
    InvalidArgument = 3,
    DeadlineExceeded = 4,
    NotFound = 5,
    AlreadyExists = 6,
    PermissionDenied = 7,
    ResourceExhausted = 8,
    FailedPrecondition = 9,
    Aborted = 10,
    OutOfRange = 11,
    Unimplemented = 12,
    Internal = 13,
    Unavailable = 14,
    DataLoss = 15,
    Unauthenticated = 16,
}

unsafe impl ZeroCopyEnum for Code {
    const MAX_VALUE: u8 = 16;
}

#[repr(C)]
pub struct Status {
    pub code: Enum<Code>,
    pub message: Str,
}

unsafe impl ZeroCopy for Status {}
