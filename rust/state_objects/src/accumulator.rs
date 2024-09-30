pub struct Accumulator {}

pub struct AccumulatorMap<K> {
    _phantom: std::marker::PhantomData<K>,
}