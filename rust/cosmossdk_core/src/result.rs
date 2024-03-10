extern crate core;
use crate::error::Error;

pub type Result<T> = core::result::Result<T, Error>;