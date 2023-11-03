use std::fmt::Write;
use crate::opts::Options;

#[derive(Default)]
pub(crate) struct Context {
    pub(crate) str: String,
    pub(crate) opts: Options,
}

impl Write for Context {
    fn write_str(&mut self, s: &str) -> std::fmt::Result {
        self.str.write_str(s)
    }
}
