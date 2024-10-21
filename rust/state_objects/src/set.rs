//! The set module contains the `Set` struct, which represents a set of keys in storage.

use crate::Map;

/// A set of keys in storage.
pub struct Set<K> {
    map: Map<K, ()>
}