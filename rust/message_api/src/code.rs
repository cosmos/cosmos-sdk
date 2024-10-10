//! Error and success codes returned by the message API.

use core::fmt::{Debug};
use num_enum::{IntoPrimitive, TryFromPrimitive};

/// Error and success codes returned by the message API.
#[derive(Clone, Copy, Debug)]
#[non_exhaustive]
pub enum ErrorCode<E: HandlerCode = u8> {
    /// A known system error code.
    SystemCode(SystemCode),

    /// A custom error code returned by a handler.
    HandlerCode(E),

    /// Unknown error code.
    Unknown(u16),
}

/// A trait implemented by all types that can be used as custom handler error codes.
pub trait HandlerCode: Into<u8> + TryFrom<u8> + Debug + Clone {}
impl<T: Into<u8> + TryFrom<u8> + Debug + Clone> HandlerCode for T {}

/// A known system error code.
#[derive(Clone, Copy, PartialEq, Eq, Debug, IntoPrimitive, TryFromPrimitive)]
#[repr(u8)]
#[non_exhaustive]
pub enum SystemCode {
    // System restricted error codes:

    /// Fatal execution error that likely cannot be recovered from.
    FatalExecutionError = 1,
    /// Account not-found error.
    AccountNotFound = 2,
    /// Message handler not-found error.
    HandlerNotFound = 3,
    /// The caller attempted to impersonate another caller and was not authorized.
    UnauthorizedCallerAccess = 4,
    /// The handler code was invalid, failed to execute properly within its virtual machine
    /// or otherwise behaved incorrectly.
    InvalidHandler = 5,

    // Known errors that can be returned by handlers or the system:

    /// Any uncategorized error.
    Other = 128,
    /// The handler doesn't handle the specified message.
    MessageNotHandled = 129,
    /// Encoding error.
    EncodingError = 130,
    /// Out of gas error.
    OutOfGas = 131,
}

impl<E: HandlerCode> From<u16> for ErrorCode<E> {
    fn from(value: u16) -> Self {
        match value {
            0..256 => {
                if let Ok(e) = SystemCode::try_from(value as u8) {
                    ErrorCode::SystemCode(e)
                } else {
                    ErrorCode::Unknown(value)
                }
            }
            256..512 => {
                if let Ok(e) = E::try_from((value - 256) as u8) {
                    ErrorCode::HandlerCode(e)
                } else {
                    ErrorCode::Unknown(value)
                }
            }
            _ => ErrorCode::Unknown(value),
        }
    }
}


impl<E: HandlerCode> Into<u16> for ErrorCode<E> {
    fn into(self) -> u16 {
        match self {
            ErrorCode::SystemCode(e) => e as u16,
            ErrorCode::HandlerCode(e) => e.into() as u16 + 256,
            ErrorCode::Unknown(e) => e,
        }
    }
}

impl<E: HandlerCode> PartialEq<Self> for ErrorCode<E> {
    fn eq(&self, other: &Self) -> bool {
        let a: u16 = self.clone().into();
        let b: u16 = other.clone().into();
        a == b
    }
}

impl<E: HandlerCode> Eq for ErrorCode<E> {}

impl SystemCode {
    /// Returns `true` if the code is a valid code for a handler to return directly,
    /// or `false` if the code is in the reserved system range.
    pub fn valid_handler_code(&self) -> bool {
        let code: u8 = (*self).into();
        code >= 128
    }
}

