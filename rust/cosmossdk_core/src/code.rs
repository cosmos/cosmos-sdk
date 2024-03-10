extern crate core;
extern crate num_enum;
use core::convert::{From};
use num_enum::{IntoPrimitive, TryFromPrimitive};

#[repr(u8)]
#[derive(Clone, Copy, IntoPrimitive, TryFromPrimitive, Eq, PartialEq, Debug)]
pub enum Code {
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