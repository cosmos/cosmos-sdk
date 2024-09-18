use alloc::borrow::Cow;
use core::fmt::{Debug, Display, Formatter};
use crate::Context;

pub struct Error {
    #[cfg(feature = "std")]
    msg: String,
    // TODO no std version
}

impl Error {
    pub fn new(ctx: &Context, msg: String) -> Error {
        Error {
            #[cfg(feature = "std")]
            msg
        }
    }
}

impl Debug for Error {
    fn fmt(&self, f: &mut Formatter<'_>) -> core::fmt::Result {
        write!(f, "{}", self.msg)
    }
}

impl Display for Error {
    fn fmt(&self, f: &mut Formatter<'_>) -> core::fmt::Result {
        write!(f, "{}", self.msg)
    }
}

impl core::error::Error for Error {}

impl<E: core::error::Error> From<E> for Error {
    fn from(value: E) -> Self {
        Error {
            #[cfg(feature = "std")]
            msg: value.to_string(),
        }
    }
}

macro_rules! fmt_error {
    ($context:ident, $($arg:tt)*) => {
        $crate::error::Error::new(context, core::format!($($arg)*))
    };
}

macro_rules! bail {
    ($($arg:tt)*) => {
        return core::result::Err($crate::error::fmt_error!($($arg)*));
    };
}

macro_rules! ensure {
    ($cond:expr, $($arg:tt)*) => {
        if !$cond {
            return core::result::Err($crate::error::fmt_error!($($arg)*));
        }
    };
}
