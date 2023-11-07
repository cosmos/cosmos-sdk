use crate::root::RawRoot;
use crate::status::Status;
use crate::{Root, ZeroCopy};

type Handler<Ctx, Err> = fn(Ctx, &RawRoot, &mut RawRoot) -> Result<(), Err>;

trait ServiceRegistrar<Ctx, Err> {
    fn register_unary(&mut self, name: &'static str, handler: Handler<Ctx, Err>);
}

trait Server<Ctx, Err> {
    const NAME: &'static str;
    fn register(&self, registrar: &mut dyn ServiceRegistrar<Ctx, Err>);
}

struct Responder<Res> {
    root: RawRoot,
    _phantom: core::marker::PhantomData<Res>,
}

impl<Res: ZeroCopy> Responder<Res> {
    fn success(self) -> Root<Res> {
        todo!()
    }

    fn error(self) -> Root<Status> {
        todo!()
    }
}
