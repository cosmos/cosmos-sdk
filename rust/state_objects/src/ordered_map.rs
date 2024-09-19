use ixc_core::{Context, Response};
use ixc_schema::state_object::{ObjectKey, ObjectValue, PrefixKey};
use crate::Map;

/// An ordered map is a map that maintains the order of its keys.
pub struct OrderedMap<K, V> {
    map: Map<K, V>,
}

impl<K: ObjectKey, V: ObjectKey> OrderedMap<K, V> {
    /// Iterate over the keys and values in the map in order.
    pub fn iterate<'a, Start, End>(&self, ctx: &Context, start: Start::Value<'_>, end: End::Value<'_>) -> Response<Iter<'a, K, V>>
    where
        Start: PrefixKey<K>,
        End: PrefixKey<K>,
    {
        todo!()
    }

    /// Iterate over the keys and values in the map in reverse order.
    pub fn iterate_reverse<'a, Start, End>(&self, ctx: &Context, start: Start::Value<'_>, end: End::Value<'_>) -> Response<Iter<'a, K, V>>
    where
        Start: PrefixKey<K>,
        End: PrefixKey<K>,
    {
        todo!()
    }
}

/// An iterator over the keys and values in an ordered map.
pub struct Iter<'a, K, V> {
    _phantom: std::marker::PhantomData<(&'a K, &'a V)>,
    _phantom2: std::marker::PhantomData<&'a ()>,
}

impl<'a, K: ObjectKey, V: ObjectValue> Iterator for Iter<'a, K, V> {
    type Item = (K::Value<'a>, V::Value<'a>);

    fn next(&mut self) -> Option<Self::Item> {
        todo!()
    }
}
