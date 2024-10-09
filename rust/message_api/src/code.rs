//! Error and success codes returned by the message API.

use core::fmt::{Debug};
use num_enum::{IntoPrimitive, TryFromPrimitive};

/// Error and success codes returned by the message API.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
#[non_exhaustive]
pub enum ErrorCode<E: Into<u8> + TryFrom<u8> + Debug = u8> {
    /// A known system error code.
    SystemCode(SystemCode),

    /// A custom error code returned by a handler.
    HandlerCode(E),

    /// Unknown error code.
    Unknown(u16),
}

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

impl<E: Into<u8> + TryFrom<u8> + Debug> From<u16> for ErrorCode<E> {
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


impl<E: Into<u8> + TryFrom<u8> + Debug> Into<u16> for ErrorCode<E> {
    fn into(self) -> u16 {
        match self {
            ErrorCode::SystemCode(e) => e as u16,
            ErrorCode::HandlerCode(e) => e.into() as u16 + 256,
            ErrorCode::Unknown(e) => e,
        }
    }
}

impl SystemCode {
    /// Returns `true` if the code is a valid code for a handler to return directly,
    /// or `false` if the code is in the reserved system range.
    pub fn valid_handler_code(&self) -> bool {
        let code: u8 = (*self).into();
        code >= 128
    }
}

// impl From<u32> for Code {
//     fn from(value: u32) -> Self {
//         match value {
//             0 => Code::Ok,
//             1 => Code::OutOfGas,
//             2 => Code::FatalExecutionError,
//             3 => Code::AccountNotFound,
//             4 => Code::MessageHandlerNotFound,
//             5 => Code::InvalidStateAccess,
//             6 => Code::UnauthorizedCallerAccess,
//             7 => Code::InvalidHandler,
//             8 => Code::UnknownHandlerError,
//             ..=255 => Code::UnknownSystemError(value),
//             _ => Code::HandlerError(value),
//         }
//     }
// }
//
// impl Into<u32> for Code {
//     fn into(self) -> u32 {
//         match self {
//             Code::Ok => 0,
//             Code::OutOfGas => 1,
//             Code::FatalExecutionError => 2,
//             Code::AccountNotFound => 3,
//             Code::MessageHandlerNotFound => 4,
//             Code::InvalidStateAccess => 5,
//             Code::UnauthorizedCallerAccess => 6,
//             Code::InvalidHandler => 7,
//             Code::UnknownHandlerError => 8,
//             Code::UnknownSystemError(value) => value,
//             Code::HandlerError(value) => value,
//         }
//     }
// }

