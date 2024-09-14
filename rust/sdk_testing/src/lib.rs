use interchain_core::{Address, Context, Response};
use interchain_core::handler::{AccountAPI, AccountHandler, ModuleAPI, ModuleHandler};

#[derive(Default)]
pub struct TestApp {}

impl TestApp {
    pub fn add_module<H: ModuleHandler>(&mut self, init: H::Init) -> Response<H> {
        todo!()
    }

    pub fn add_mock_module(&mut self, module_name: &str) -> MockModule {
        todo!()
    }

    pub fn add_account<H: AccountHandler>(&mut self, ctx: &mut Context, init: H::Init) -> Response<(Address, H)> {
        todo!()
    }

    pub fn add_mock_account(&mut self, ctx: &mut Context) -> Response<(Address, MockAccount)> {
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


pub struct MockModule<'a> { }

impl MockModule {
    fn add_mock_module_api<A: ModuleAPI>(&mut self, mock: &A) {
        todo!()
    }

    fn add_mock_account_api<A: ModuleAPI>(&mut self, mock: &A) {
        todo!()
    }
}

pub struct MockAccount<'a> {}

impl MockAccount {
    fn address(&self) -> &Address {
        todo!()
    }

    fn add_mock_account_api<A: ModuleAPI>(&mut self, mock: &A) {
        todo!()
    }
}