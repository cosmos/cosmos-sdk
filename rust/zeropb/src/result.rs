use core::result;
use cosmossdk_core::Code;
use crate::{Error, Root, Str, ZeroCopy};

pub type RawResult<T> = result::Result<T, Error>;
pub type Result<T> = RawResult<Root<T>>;

pub fn ok<T: ZeroCopy>() -> Result<T> {
    Ok(Root::empty())
}

pub fn err_code<T: ZeroCopy>(code: Code) -> Result<T> {
    Err(Error {
        code,
        msg: Root::empty(),
    })
}

pub fn err_msg<T: ZeroCopy>(code: Code, message: &str) -> Result<T> {
    let mut msg = Root::<Str>::new();
    msg.set(message)?;
    Err(Error {
        code,
        msg,
    })
}

pub fn err_code_raw<T>(code: Code) -> RawResult<T> {
    Err(Error {
        code,
        msg: Root::empty(),
    })
}
