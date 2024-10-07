//! Field definition.
use crate::kind::Kind;

/// A field in a type.
#[non_exhaustive]
#[derive(Debug, Clone, Copy, Eq, PartialEq)]
pub struct Field<'a> {
    pub name: &'a str,
    pub kind: Kind,
    pub nullable: bool,
    pub element_kind: Option<Kind>,
    pub referenced_type: &'a str,
}

impl <'a> Field<'a> {
    /// Returns a copy of the field with the provided name set.
    pub const fn with_name(mut self, name: &'a str) -> Self {
        self.name = name;
        self
    }
}

