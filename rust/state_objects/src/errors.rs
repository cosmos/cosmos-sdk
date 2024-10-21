use thiserror::Error;

#[derive(Error, Debug)]
pub enum SetError {
    #[error("unknown error: {0}")]
    Unknown(String),
}

#[derive(Error, Debug)]
pub enum GetError {
    #[error("not found")]
    NotFound,
    #[error("unknown error: {0}")]
    Unknown(String),
}