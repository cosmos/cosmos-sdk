#![doc = include_str!(concat!(env!("CARGO_MANIFEST_DIR"), "/README.md"))]

mod store;
mod vm;

use std::cell::RefCell;
use std::sync::{Arc, RwLock};
use allocator_api2::alloc::Allocator;
use ixc_message_api::{AccountID};
use ixc_core::{Context};
use ixc_core::account_api::create_account;
use ixc_core::handler::{HandlerAPI, Handler, ClientFactory, Client};
use ixc_core::resource::{InitializationError, ResourceScope, Resources};
use ixc_core::routes::{Route, Router};
use ixc_hypervisor::Hypervisor;
use ixc_message_api::code::ErrorCode;
use ixc_message_api::handler::{HandlerError, HandlerErrorCode, HostBackend, RawHandler};
use ixc_message_api::header::MessageHeader;
use ixc_message_api::packet::MessagePacket;
use ixc_schema::mem::MemoryManager;
use crate::store::{Store, VersionedMultiStore};
use crate::vm::{NativeVM, NativeVMImpl};

/// Defines a test harness for running tests against account and module implementations.
pub struct TestApp {
    hypervisor: RefCell<Hypervisor<VersionedMultiStore>>,
    native_vm: NativeVM,
    mem: MemoryManager
}

impl Default for TestApp {
    fn default() -> Self {
        let mut hypervisor: Hypervisor<VersionedMultiStore> = Default::default();
        let native_vm = NativeVM::new();
        hypervisor.register_vm("native", Box::new(native_vm.clone())).unwrap();
        hypervisor.set_default_vm("native").unwrap();
        let mem = MemoryManager::new();
        let mut test_app = Self {
            hypervisor: RefCell::new(hypervisor),
            native_vm,
            mem,
        };
        test_app.register_handler::<DefaultAccount>().unwrap();
        test_app
    }
}

struct DefaultAccount;
struct DefaultAccountClient(AccountID);

unsafe impl Router for DefaultAccount { const SORTED_ROUTES: &'static [Route<Self>] = &[]; }

unsafe impl Resources for DefaultAccount {
    unsafe fn new(scope: &ResourceScope) -> Result<Self, InitializationError> {
        Ok(DefaultAccount {})
    }
}

impl ClientFactory for DefaultAccount {
    type Client = DefaultAccountClient;

    fn new_client(account_id: AccountID) -> Self::Client {
        DefaultAccountClient(account_id)
    }
}

impl Client for DefaultAccountClient {
    fn account_id(&self) -> AccountID {
        self.0
    }
}

impl Handler for DefaultAccount {
    const NAME: &'static str = "ixc_testing.DefaultAccount";
    type Init<'a> = ();
    type InitCodec = ixc_schema::binary::NativeBinaryCodec;
}

impl RawHandler for DefaultAccount {
    fn handle(&self, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend, allocator: &dyn Allocator) -> Result<(), HandlerError> {
        ixc_core::routes::exec_route(self, message_packet, callbacks, allocator)
    }
}

impl TestApp {
    /// Registers a handler with the test harness so that accounts backed by this handler can be created.
    pub fn register_handler<H: Handler>(&mut self) -> core::result::Result<(), InitializationError>{
        let scope = ResourceScope::default();
        unsafe { self.native_vm.register_handler::<H>(H::NAME, H::new(&scope)?); }
        Ok(())
    }
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
    // /// Adds an account to the test harness.
    // pub fn create_account<'a, H: Handler>(&self, ctx: &mut Context, init: &H::Init<'a>) -> Result<AccountInstance<H>, ()> {
    //     // self.native_vm.register_handler()
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

    /// Creates a new random client account that can be used in calls.
    pub fn new_client_account(&mut self) -> core::result::Result<AccountID, ()> {
        let mut ctx = self.client_context_for(AccountID::new(1));
        let client = create_account::<DefaultAccount>(&mut ctx, &())
            .map_err(|_| ())?;
        Ok(client.0)
    }

    /// Creates a new client for the given account.
    pub fn client_context_for(&mut self, account_id: AccountID) -> Context
    {
        let packet = self.mem.allocate_packet(0).unwrap();
        packet.header_mut().account = account_id;
        let ctx = Context::new(packet, self);
        ctx
    }

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

impl HostBackend for TestApp {
    fn invoke(&self, message_packet: &mut MessagePacket, allocator: &dyn Allocator) -> Result<(), ErrorCode> {
        self.hypervisor.borrow_mut().invoke(message_packet, allocator)
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