#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum SystemErrorCode {
    OutOfGas,
    FatalExecutionError,
    AccountNotFound,
    MessageHandlerNotFound,
    InvalidStateAccess,
    UnauthorizedCallerAccess,
    InvalidHandler,
    UnknownHandlerError,
    Unknown(u32),
}

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum Code {
    Ok,
    SystemError(SystemErrorCode),
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

