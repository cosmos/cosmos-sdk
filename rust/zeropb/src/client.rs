use crate::root::RawRoot;

pub type Handler<'a, Ctx, Err> = fn(Ctx, &RawRoot, &mut RawRoot) -> Result<(), Err>;

pub trait ClientConn<'a, Ctx, Err> {
    fn resolve_unary(&self, name: &'static str) -> Handler<'a, Ctx, Err>;
}

#[cfg(test)]
mod tests {}
