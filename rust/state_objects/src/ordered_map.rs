use interchain_core::{Context, Response};
use crate::codec::{ObjectKey, ObjectValue, PrefixKey};
use crate::Map;

pub struct OrderedMap<K, V> {
    map: Map<K, V>,
}

impl<'a, K: ObjectKey<'a>, V: ObjectKey<'a>> OrderedMap<K, V> {
    pub fn iterate<Start, End>(&self, ctx: &'a Context, start: Start::Value, end: End::Value) -> Response<Iter<'a, K, V>>
    where
        Start: PrefixKey<'a, K>,
        End: PrefixKey<'a, K>,
    {
        todo!()
    }

    pub fn iterate_reverse<Start, End>(&self, ctx: &Context, start: Start::Value, end: End::Value) -> Response<Iter<'a, K, V>>
    where
        Start: PrefixKey<'a, K>,
        End: PrefixKey<'a, K>,
    {
        todo!()
    }
}

pub struct Iter<'a, K, V> {
    _phantom: std::marker::PhantomData<(&'a K, &'a V)>,
    _phantom2: std::marker::PhantomData<&'a ()>,
}

impl<'a, K: ObjectKey<'a>, V: ObjectValue<'a>> Iterator for Iter<'a, K, V> {
    type Item = (K::Value, V::Value);

    fn next(&mut self) -> Option<Self::Item> {
        todo!()
    }
}
