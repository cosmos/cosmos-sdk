use crate::root::RawRoot;

type Handler<Ctx, Err> = fn(Ctx, &RawRoot, &mut RawRoot) -> Result<(), Err>;

trait ServiceRegistrar<Ctx, Err> {
    fn register_unary(&mut self, name: &'static str, handler: Handler<Ctx, Err>);
}

trait Server<Ctx, Err> {
    const NAME: &'static str;
    fn register(&self, registrar: &mut dyn ServiceRegistrar<Ctx, Err>);
}