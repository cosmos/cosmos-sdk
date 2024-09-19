//! Basic error handling utilities.
use core::fmt::{Debug, Display, Formatter};
use interchain_schema::types::StrT;
use interchain_schema::value::{MaybeBorrowed, Value};

/// A simple error type which just contains an error message.
#[derive(Clone)]
pub struct ErrorMessage {
    #[cfg(feature = "std")]
    msg: String,
    // TODO no std version
}

impl ErrorMessage {
    fn new(value: &str) -> Self {
        ErrorMessage {
            #[cfg(feature = "std")]
            msg: value.to_string(),
        }
    }
}

impl Debug for ErrorMessage {
    fn fmt(&self, f: &mut Formatter<'_>) -> core::fmt::Result {
        write!(f, "{}", self.msg)
    }
}

impl Display for ErrorMessage {
    fn fmt(&self, f: &mut Formatter<'_>) -> core::fmt::Result {
        write!(f, "{}", self.msg)
    }
}

impl<E: core::error::Error> From<E> for ErrorMessage {
    fn from(value: E) -> Self {
        ErrorMessage {
            #[cfg(feature = "std")]
            msg: value.to_string(),
        }
    }
}

impl <'a> MaybeBorrowed<'a> for ErrorMessage {
    type Type = StrT;
}

impl Value for ErrorMessage {
    type MaybeBorrowed<'a> = ErrorMessage;
}

/// Format an error message.
#[macro_export]
macro_rules! fmt_error {
    ($($arg:tt)*) => {
        $crate::error:ErrorMessage::new(core::format!($($arg)*))
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
