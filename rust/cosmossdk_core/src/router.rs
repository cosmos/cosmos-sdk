pub trait Router {
    fn route(&self, method_id: u64, ctx: u64, p0: usize, p1: usize) -> usize;
}