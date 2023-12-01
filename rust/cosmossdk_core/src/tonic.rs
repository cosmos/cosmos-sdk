#![cfg(feature = "tonic")]

use tonic::Request;
use crate::context::Context;

pub fn context<T>(req: &Request<T>) -> Context {
    // get context from the context metadata key encoded as an int string
    let ctx_id = req
        .metadata()
        .get("context")
        .expect("context metadata key not found")
        .to_str()
        .expect("context metadata value is not a string")
        .parse::<u32>()
        .expect("context metadata value is not an int");
    Context { id: ctx_id }
}
pub fn mut_context<T>(req: &Request<T>) -> Context {
    // get context from the mut_context metadata key encoded as an int string
    let ctx_id = req
        .metadata()
        .get("mut_context")
        .expect("mut_context metadata key not found")
        .to_str()
        .expect("mut_context metadata value is not a string")
        .parse::<u32>()
        .expect("mut_context metadata value is not an int");
    Context { id: ctx_id }
}
