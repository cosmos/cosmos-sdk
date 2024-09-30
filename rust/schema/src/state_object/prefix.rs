use crate::state_object::key::ObjectKey;

/// This trait is implemented for types that can be used as prefix keys in state objects.
pub trait PrefixKey<K: ObjectKey> {
    /// The possibly borrowed value type to use.
    type Value<'a>;
}
