use ixc_core::Context;
use ixc_message_api::handler::Handler;

pub struct HandlerID {
    vm: u32,
    vm_handler_id: VMHandlerID,
}

pub struct VMHandlerID {
    package: u64,
    handler: u32,
}

pub trait VM {
    fn get_handler(&self, ctx: &Context, vm_handler_id: VMHandlerID) -> &dyn Handler;
}