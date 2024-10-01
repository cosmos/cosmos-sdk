//! Error and success codes returned by the message API.

use crate::handler::HandlerErrorCode;

/// Error codes that can be returned by the system.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum SystemErrorCode {

    // System restricted error codes:

    /// Fatal execution error that likely cannot be recovered from.
    FatalExecutionError = 1,
    /// Account not-found error.
    AccountNotFound = 2,
    /// Message handler not-found error.
    HandlerNotFound = 3,
    /// The caller attempted to impersonate another caller and was not authorized.
    UnauthorizedCallerAccess = 4,
    /// The handler code was invalid or failed to execute properly within its virtual machine.
    InvalidHandler = 5,
    /// The handler returned an invalid error code.
    UnknownHandlerError = 6,

    // System errors that can be returned by handlers:

    /// The handler doesn't handle the specified message.
    MessageNotHandled = 128,
    /// Encoding error.
    EncodingError = 129,
    /// Out of gas error.
    OutOfGas = 130,

    /// An unknown error code in the system range.
    Unknown(u32),
}

/// Error and success codes returned by the message API.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum ErrorCode {
    /// An error that can only be returned by the system, in the range of 0..127.
    RuntimeSystemError(SystemErrorCode),
    /// A predefined error code returned by handler implementations libraries,
    /// in the range of 128..255.
    HandlerSystemError(HandlerErrorCode),
    /// A custom error code returned by a handler.
    CustomHandlerError(u16)
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

// pub type Handler = unsafe fn(account_handler_id: u64, message_packet: *mut u8, packet_len: u32) -> u32;
//
// pub type InvokeFn = unsafe fn(message_packet: *mut u8, packet_len: u32) -> u32;

