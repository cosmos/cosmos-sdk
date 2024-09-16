use interchain_core::{Context, Response};
use crate::Item;

pub struct Seq {
    value: Item<u64>,
}

impl Seq {
    pub fn peek(&self, ctx: &Context) -> Response<u64> {
        todo!()
    }

    pub fn next(&self, ctx: &mut Context) -> Response<u64> {
        todo!()
    }

    pub unsafe fn set(&self, ctx: &mut Context, value: u64) -> Response<()> {
        todo!()
    }
}