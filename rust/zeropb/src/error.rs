use thiserror::Error;

#[derive(Error, Debug)]
pub enum Error {
    #[error("out of memory")]
    OutOfMemory,

    #[error("out of bounds")]
    OutOfBounds,

    #[error("invalid state")]
    InvalidState,
}

