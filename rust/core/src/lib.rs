#![doc = include_str!(concat!(env!("CARGO_MANIFEST_DIR"), "/README.md"))]
#![no_std]

#[cfg(feature = "std")]
extern crate alloc;

mod context;
mod events;
pub mod message;
pub mod account_api;
pub mod handler;
pub mod resource;
pub mod error;
pub mod routing;
pub mod low_level;
pub mod result;

pub use context::Context;
pub use events::EventBus;
pub use handler::Service;
pub use account_api::create_account;

pub use result::{Result};

/// Create an error.
#[macro_export]
macro_rules! error {
    ($code:path) => {
        $crate::error::HandlerError::new_from_code($code)
    };
    ($code:path, $str:literal, $($arg:tt)*) => {
        $crate::error::HandlerError::new_fmt_with_code($code, core::format_args!($str, $($arg)*))
    };
    ($str:literal) => {
        $crate::error::HandlerError::new($str.to_string())
    };
    ($str:literal, $($arg:tt)*) => {
        $crate::error::HandlerError::new_fmt(core::format_args!($str, $($arg)*))
    };
}

/// Return an error with a formatted message.
#[macro_export]
macro_rules! bail {
    ($($arg:tt)*) => {
        return core::result::Result::Err($crate::error!($($arg)*));
    };
}

/// Ensure a condition is true, otherwise return an error with a formatted message.
#[macro_export]
macro_rules! ensure {
    ($cond:expr, $($arg:tt)*) => {
        if !$cond {
            return core::result::Result::Err($crate::error!($($arg)*));
        }
    };
}
