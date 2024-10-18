use crate::kind::Kind;

#[non_exhaustive]
#[derive(Debug, Clone, Eq, PartialEq)]
pub struct OneOfType<'a> {
    pub name: &'a str,
    pub cases: &'a [OneOfCase<'a>],
    pub discriminant_kind: Kind,
    pub sealed: bool,
}

#[non_exhaustive]
#[derive(Debug, Clone, Eq, PartialEq)]
pub struct OneOfCase<'a> {
    pub name: &'a str,
    pub discriminant: i32,
    pub kind: Kind,
    pub referenced_type: &'a str,
}