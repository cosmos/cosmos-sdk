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
        .parse::<usize>()
        .expect("context metadata value is not an int");
    Context { id: ctx_id }
}
