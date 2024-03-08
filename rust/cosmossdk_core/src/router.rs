pub trait Router {
    fn route(&self, route_id: u64, ctx: usize, p0: usize, p1: usize) -> usize;
}