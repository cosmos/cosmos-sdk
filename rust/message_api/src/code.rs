//! Error and success codes returned by the message API.

/// Error codes that can be returned by the system.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum SystemErrorCode {
    /// Out of gas error.
    OutOfGas,
    /// Fatal execution error that likely cannot be recovered from.
    FatalExecutionError,
    /// Account not-found error.
    AccountNotFound,
    /// The caller attempted to impersonate another caller and was not authorized.
    UnauthorizedCallerAccess,
    /// The handler code was invalid or failed to execute properly within its virtual machine.
    InvalidHandler,
    /// The handler returned an invalid error code.
    UnknownHandlerError,
    /// The system encountered an unknown error.
    Unknown(u32),
}

/// Error and success codes returned by the message API.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum Code {
    /// The operation completed successfully.
    Ok,
    /// A system error.
    SystemError(SystemErrorCode),
    /// An error returned by the handler.
    HandlerError(u32),
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

