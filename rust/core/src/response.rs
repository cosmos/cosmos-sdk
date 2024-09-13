/// Response is the type that should be used for message handler responses.
pub type Response<R, E=()> = Result<R, E>;
