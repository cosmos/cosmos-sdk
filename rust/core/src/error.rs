//! Basic error handling utilities.

use alloc::string::{String, ToString};
use core::fmt::{Debug, Display, Formatter};
use ixc_message_api::code::SystemErrorCode;
use ixc_message_api::handler::HandlerErrorCode;
use ixc_schema::decoder::{DecodeError, Decoder};
use ixc_schema::encoder::{EncodeError, Encoder};
use ixc_schema::mem::MemoryManager;
use ixc_schema::types::StrT;
use ixc_schema::value::{SchemaValue, OptionalValue};

/// The standard error wrapper for handler functions.
#[derive(Debug, Clone)]
pub enum Error<E: OptionalValue<'static>> {
    /// A system error occurred.
    SystemError(SystemErrorCode),
    /// A known handler error occurred.
    KnownHandlerError(HandlerErrorCode),
    /// A custom handler error occurred.
    HandlerError(E), // TODO response body
}

/// A simple error type which just contains an error message.
#[derive(Clone)]
pub struct ErrorMessage {
    #[cfg(feature = "std")]
    msg: String,
    // TODO no std version - fixed length 256 byte string probably
}

impl ErrorMessage {
    fn new(msg: String) -> Self {
        ErrorMessage {
            #[cfg(feature = "std")]
            msg,
        }
    }

    fn new_fmt(args: core::fmt::Arguments<'_>) -> Self {
        #[cfg(feature = "std")]
        let mut message = String::new();
        core::fmt::write(&mut message, args).unwrap();
        ErrorMessage::new(message)
    }
}

impl<'a> Debug for ErrorMessage {
    fn fmt(&self, f: &mut Formatter<'_>) -> core::fmt::Result {
        write!(f, "{}", self.msg)
    }
}

impl<'a> Display for ErrorMessage {
    fn fmt(&self, f: &mut Formatter<'_>) -> core::fmt::Result {
        write!(f, "{}", self.msg)
    }
}

impl<'a, E: core::error::Error> From<E> for ErrorMessage {
    fn from(value: E) -> Self {
        ErrorMessage {
            #[cfg(feature = "std")]
            msg: value.to_string(),
        }
    }
}

impl<'a> SchemaValue<'a> for ErrorMessage {
    type Type = StrT;
    type DecodeState = String;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> core::result::Result<(), DecodeError> {
        *state = decoder.decode_owned_str()?;
        Ok(())
    }

    fn finish_decode_state(msg: Self::DecodeState, _mem_handle: &'a MemoryManager) -> core::result::Result<Self, DecodeError> {
        Ok(ErrorMessage { msg })
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> core::result::Result<(), EncodeError> {
        encoder.encode_str(&self.msg)
    }

}

/// Format an error message.
#[macro_export]
macro_rules! fmt_error {
    ($($arg:tt)*) => {
        $crate::error::ErrorMessage::new(core::format_args!($($arg)*))
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
