//! Field definition.
use crate::kind::Kind;

/// A field in a type.
#[non_exhaustive]
#[derive(Debug, Clone, Copy, Eq, PartialEq)]
pub struct Field<'a> {
    /// The name of the field.
    pub name: &'a str,
    /// The kind of the field.
    pub kind: Kind,
    /// Whether the field is nullable.
    pub nullable: bool,
    /// The element kind for list fields.
    pub element_kind: Option<Kind>,
    /// The referenced type for fields which reference another type.
    pub referenced_type: &'a str,
}

impl <'a> Field<'a> {
    /// Returns a copy of the field with the provided name set.
    pub const fn with_name(mut self, name: &'a str) -> Self {
        self.name = name;
        self
    }
}

