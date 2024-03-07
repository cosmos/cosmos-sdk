use zeropb::Context;

use crate::router::Router;

trait Handler<Request, Response = ()> {
    fn handle(&self, ctx: &mut Context, req: &Request) -> zeropb::Result<Response>;
}

trait InternalHandler<Request, Response = ()> {
    fn handle(&self, ctx: &mut Context, caller_id: &zeropb::ModuleID, req: &Request) -> zeropb::Result<Response>;
}

trait EventHook<Event> {
    fn on_event(&self, ctx: &mut Context, event: &Event) -> zeropb::Result<()>;
}

trait PreHandler<Request> {
    fn pre_handle(&self, ctx: &mut Context, req: &Request) -> zeropb::Result<()>;
}

trait PostHandler<Request, Response = ()> {
    fn post_handle(&self, ctx: &mut Context, req: &Request, res: &mut Response) -> zeropb::Result<()>;
}

impl<T, Request, Response> Router for T
    where T: Handler<Request, Response>
{
    fn route(&self, method_id: u64, ctx: usize, caller_id: u64, p0: usize, p1: usize) -> usize {
        let res = self.handle(todo!(), todo!());
        0
    }
}

impl<T, Request, Response> Router for T
    where T: InternalHandler<Request, Response>
{
    fn route(&self, method_id: u64, ctx: usize, caller_id: u64, p0: usize, p1: usize) -> usize {
        let res = self.handle(todo!(), todo!(), todo!());
        0
    }
}
