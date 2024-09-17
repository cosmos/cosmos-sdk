use crate::OrderedMap;

/// An ordered set of keys in storage.
pub struct OrderedSet<K> {
    map: OrderedMap<K, ()>
}