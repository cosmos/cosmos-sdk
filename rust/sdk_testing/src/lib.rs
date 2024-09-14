use interchain_core::{Address, Context, Response};
use interchain_core::handler::{AccountHandler, ModuleHandler};

#[derive(Default)]
pub struct TestApp {}

impl TestApp {
    pub fn add_module<H: ModuleHandler>(init: H::Init) -> Response<H> {
        todo!()
    }

    pub fn new_account<H: AccountHandler>(ctx: &mut Context, init: H::Init) -> Response<(Address, H)> {
        todo!()
    }

    pub fn new_client_context(&mut self) -> Context {
        todo!()
    }

    pub fn storage(&self) -> &TestStorage {
        todo!()
    }

    pub fn storage_mut(&mut self) -> &mut TestStorage {
        todo!()
    }
}

pub struct TestStorage {}

impl TestStorage {
    pub fn begin_tx(&mut self, ctx: &Context) -> Response<Context> {
        todo!()
    }

    pub fn rollback_tx(&mut self, ctx: &mut Context) {
        todo!()
    }

    pub fn commit_tx(&mut self, ctx: &mut Context) {
        todo!()
    }
}

