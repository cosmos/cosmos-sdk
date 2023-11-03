use crate::root::RawRoot;

type Handler<Ctx, Err> = fn(Ctx, &RawRoot, &mut RawRoot) -> Result<(), Err>;

pub trait ClientConn<Ctx, Err> {
    fn resolve_unary(&mut self, name: &'static str) -> &Handler<Ctx, Err>;
}

#[cfg(test)]
mod tests {

}