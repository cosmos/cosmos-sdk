use crate::{Bytes, Str};

pub enum ModuleID {
    Module{name: Str},
    Account{address: Bytes},
}