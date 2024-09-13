//! The item module contains the `Item` struct, which represents a single item in storage.

use crate::Map;

/// A single item in storage.
pub struct Item<V> {
    map: Map<(), V>
}