//! The map module contains the `Map` struct, which represents a key-value map in storage.

/// A key-value map.
pub struct Map<K, V> {
    _phantom: std::marker::PhantomData<(K, V)>,
}