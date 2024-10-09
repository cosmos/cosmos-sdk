//! Basic error handling utilities.

use alloc::format;
use alloc::string::String;
use core::error::Error;
use core::fmt::{Debug, Display, Formatter};
use ixc_message_api::code::{ErrorCode, SystemCode};
use ixc_schema::decoder::DecodeError;
use ixc_schema::encoder::EncodeError;

/// The standard error type returned by handlers.
#[derive(Clone)]
pub struct HandlerError<E: Into<u8> + TryFrom<u8> + Debug> {
    pub(crate) code: Option<E>,
    #[cfg(feature = "std")]
    pub(crate) msg: String,
    // TODO no std version - fixed length 256 byte string probably
}

impl<E: Into<u8> + TryFrom<u8> + Debug> HandlerError<E> {
    /// Create a new error message.
    pub fn new(msg: String) -> Self {
        HandlerError {
            code: None,
            #[cfg(feature = "std")]
            msg,
        }
    }

    /// Create a new error message with a code.
    pub fn new_with_code(code: E, msg: String) -> Self {
        HandlerError {
            code: Some(code),
            #[cfg(feature = "std")]
            msg,
        }
    }

    /// Format a new error message.
    pub fn new_fmt(args: core::fmt::Arguments<'_>) -> Self {
        #[cfg(feature = "std")]
        let mut message = String::new();
        core::fmt::write(&mut message, args).unwrap();
        HandlerError::new(message)
    }

    /// Format a new error message with a code.
    pub fn new_fmt_with_code(code: E, args: core::fmt::Arguments<'_>) -> Self {
        #[cfg(feature = "std")]
        let mut message = String::new();
        core::fmt::write(&mut message, args).unwrap();
        HandlerError::new(message)
    }
}

impl<E: Into<u8> + TryFrom<u8> + Debug> Debug for HandlerError<E> {
    fn fmt(&self, f: &mut Formatter<'_>) -> core::fmt::Result {
        if let Some(code) = &self.code {
            write!(f, "code: {:?}: {}", code, self.msg)
        } else {
            write!(f, "{}", self.msg)
        }
    }
}

impl<E: Into<u8> + TryFrom<u8> + Debug> Display for HandlerError<E> {
    fn fmt(&self, f: &mut Formatter<'_>) -> core::fmt::Result {
        if let Some(code) = &self.code {
            write!(f, "code: {:?}: {}", code, self.msg)
        } else {
            write!(f, "{}", self.msg)
        }
    }
}

impl<E: Error, F: Into<u8> + TryFrom<u8> + Debug> From<E> for HandlerError<F> {
    fn from(value: E) -> Self {
        HandlerError {
            code: None,
            msg: format!("got error: {}", value),
        }
    }
}

/// Format an error message.
#[macro_export]
macro_rules! fmt_error {
    ($code:ident, $($arg:tt)*) => {
        $crate::error::HandlerError::new_fmt_with_code($code, core::format_args!($($arg)*))
    };
    ($($arg:tt)*) => {
        $crate::error::HandlerError::new_fmt(core::format_args!($($arg)*))
    };
}

/// Return an error with a formatted message.
#[macro_export]
macro_rules! bail {
    ($($arg:tt)*) => {
        return core::result::Err($crate::error::fmt_error!($($arg)*));
    };
}

/// Ensure a condition is true, otherwise return an error with a formatted message.
#[macro_export]
macro_rules! ensure {
    ($cond:expr, $($arg:tt)*) => {
        if !$cond {
            return core::result::Err($crate::error::fmt_error!($($arg)*));
        }
    };
}

/// The standard error type returned by client methods.
#[derive(Clone)]
pub struct ClientError<E: Into<u8> + TryFrom<u8> + Debug> {
    /// The error code.
    pub code: ErrorCode<E>,
    /// The error message.
    #[cfg(feature = "std")]
    pub message: String,
    // TODO no std version - fixed length 256 byte string probably
}

impl<E: Into<u8> + TryFrom<u8> + Debug> ClientError<E> {
    /// Creates a new client error.
    pub fn new(code: ErrorCode<E>, msg: String) -> Self {
        ClientError {
            code,
            #[cfg(feature = "std")]
            message: msg,
        }
    }
}

impl<E: Into<u8> + TryFrom<u8> + Debug> Debug for ClientError<E> {
    fn fmt(&self, f: &mut Formatter<'_>) -> core::fmt::Result {
        match self.code {
            ErrorCode::SystemCode(SystemCode::Other) => write!(f, "{}", self.message),
            _ => write!(f, "code: {:?}: {}", self.code, self.message)
        }
    }
}

impl<E: Into<u8> + TryFrom<u8> + Debug> Display for ClientError<E> {
    fn fmt(&self, f: &mut Formatter<'_>) -> core::fmt::Result {
        match self.code {
            ErrorCode::SystemCode(SystemCode::Other) => write!(f, "{}", self.message),
            _ => write!(f, "code: {:?}: {}", self.code, self.message)
        }
    }
}

impl<E: Into<u8> + TryFrom<u8> + Debug> Error for ClientError<E> {}

impl<E: Into<u8> + TryFrom<u8> + Debug> From<ErrorCode> for ClientError<E> {
    fn from(value: ErrorCode) -> Self {
        let code = convert_error_code(value);
        ClientError {
            code,
            #[cfg(feature = "std")]
            message: String::new(),
        }
    }
}

impl<E: Into<u8> + TryFrom<u8> + Debug> From<EncodeError> for ClientError<E> {
    fn from(value: EncodeError) -> Self {
        ClientError {
            code: ErrorCode::SystemCode(SystemCode::EncodingError),
            #[cfg(feature = "std")]
            message: format!("encoding error: {:?}", value),
        }
    }
}

impl<E: Into<u8> + TryFrom<u8> + Debug> From<DecodeError> for ClientError<E> {
    fn from(value: DecodeError) -> Self {
        ClientError {
            code: ErrorCode::SystemCode(SystemCode::EncodingError),
            #[cfg(feature = "std")]
            message: format!("decoding error: {:?}", value),
        }
    }
}

impl<E: Into<u8> + TryFrom<u8> + Debug> From<allocator_api2::alloc::AllocError> for ClientError<E> {
    fn from(_: allocator_api2::alloc::AllocError) -> Self {
        ClientError {
            code: ErrorCode::SystemCode(SystemCode::EncodingError),
            #[cfg(feature = "std")]
            message: "allocation error".into(),
        }
    }
}

/// Converts an error code with one handler code to an error code with another handler code.
pub fn convert_error_code<E: Into<u8> + TryFrom<u8> + Debug, F: Into<u8> + TryFrom<u8> + Debug>(code: ErrorCode<E>) -> ErrorCode<F> {
    let c: u16 = code.into();
    ErrorCode::<F>::from(c)
}

/// Converts an error code with one handler code to an error code with another handler code.
pub fn convert_client_error<E: Into<u8> + TryFrom<u8> + Debug, F: Into<u8> + TryFrom<u8> + Debug>(err: ClientError<E>) -> ClientError<F> {
    ClientError {
        code: convert_error_code(err.code),
        #[cfg(feature = "std")]
        message: err.message,
    }
}
