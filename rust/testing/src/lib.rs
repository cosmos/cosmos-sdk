#![doc = include_str!(concat!(env!("CARGO_MANIFEST_DIR"), "/README.md"))]

mod store;
mod vm;

use ixc_message_api::{AccountID};
use ixc_core::{Context};
use ixc_core::handler::{HandlerAPI, Handler};
use ixc_hypervisor::Hypervisor;
use crate::store::{Store, VersionedMultiStore};

/// Defines a test harness for running tests against account and module implementations.
#[derive(Default)]
pub struct TestApp {
    hypervisor: Hypervisor<VersionedMultiStore>
}

impl TestApp {
    // /// Adds a module to the test harness.
    // pub fn add_module<H: ModuleHandler>(&mut self, module_name: &str, init: H::Init) -> Result<AccountInstance<H>, ()> {
    //     todo!()
    // }
    //
    // /// Adds a mock module to the test harness.
    // pub fn add_mock_module(&mut self, module_name: &str, mock: MockModule) {
    //     todo!()
    // }
    //
    // /// Adds a mock module to the test harness.
    // pub fn add_account<H: AccountHandler>(&mut self, caller: &Address, init: H::Init) -> Result<AccountInstance<H>, ()> {
    //     todo!()
    // }
    //
    // /// Adds a mock account to the test harness with the given address.
    // pub fn add_account_with_address<H: AccountHandler>(&mut self, caller: &Address, address: &Address, init: H::Init) -> Result<AccountInstance<H>, ()> {
    //     todo!()
    // }
    //
    // /// Adds a mock account to the test harness.
    // pub fn add_mock_account(&mut self, ctx: &mut Context, mock: MockAccount) -> Result<Address, ()> {
    //     todo!()
    // }
    //
    // /// Adds a mock account to the test harness with the given address.
    // pub fn add_mock_account_with_address(&mut self, address: &Address, mock: MockAccount) -> Result<Address, ()> {
    //     todo!()
    // }
    //
    // /// Creates a new random client address that can be used in calls.
    // pub fn new_client_address(&mut self) -> Address {
    //     todo!()
    // }
    //
    // /// Creates a new client context with a random address.
    // pub fn client_context(&mut self, address: &Address) -> &mut Context {
    //     todo!()
    // }
    // //
    // // /// Creates a new client context with the given address.
    // // pub fn new_client_context_with_address(&mut self, address: &Address) -> Context {
    // //     todo!()
    // // }

    /// Returns the test storage.
    pub fn storage(&self) -> &TestStorage {
        todo!()
    }

    /// Returns a mutable reference to the test storage.
    pub fn storage_mut(&mut self) -> &mut TestStorage {
        todo!()
    }
}

/// Defines the test storage implementation.
pub struct TestStorage {}

impl TestStorage {
    /// Begins a transaction.
    pub fn begin_tx(&mut self, ctx: &Context) -> Result<Context, ()> {
        todo!()
    }

    /// Rolls back a transaction.
    pub fn rollback_tx(&mut self, ctx: &mut Context) {
        todo!()
    }

    /// Commits a transaction.
    pub fn commit_tx(&mut self, ctx: &mut Context) {
        todo!()
    }
}

/// Defines a test account instance.
pub struct AccountInstance<'a, H: Handler> {
    _phantom: std::marker::PhantomData<&'a ()>,
    _phantom2: std::marker::PhantomData<H>,
}

impl<'a, H: Handler> AccountInstance<'a, H> {
    /// Returns the address of the account.
    fn account_id(&self) -> AccountID {
        todo!()
    }

    /// Executes the closure in the context of the account.
    /// This can be used for reading its internal state.
    fn with_context<F, R>(&self, f: F) -> R
    where
        F: FnOnce(&Context, &H) -> R,
    {
        todo!()
    }

    /// Executes the closure in the context of the account.
    /// This can be used for reading and modifying its internal state.
    fn with_context_mut<F, R>(&mut self, f: F) -> R
    where
        F: FnOnce(&Context) -> R,
    {
        todo!()
    }
}

/// Defines a mock module handler composed of mock module and account API trait implementations.
// pub struct MockModule {}
//
// impl MockModule {
//     /// Adds a mock module API implementation to the mock module handler.
//     fn add_mock_module_api<A: ModuleAPI>(&mut self, mock: A) {
//         todo!()
//     }
//
//     /// Adds a mock account API implementation to the mock module handler.
//     fn add_mock_account_api<A: HandlerAPI>(&mut self, mock: A) {
//         todo!()
//     }
// }
//

/// Defines a mock account handler composed of mock account API trait implementations.
pub struct MockAccount {}

impl MockAccount {
    /// Adds a mock account API implementation to the mock account handler.
    fn add_mock_account_api<A: HandlerAPI>(&mut self, mock: A) {
        todo!()
    }
}