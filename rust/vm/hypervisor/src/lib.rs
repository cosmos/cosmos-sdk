use std::alloc::Layout;
use std::cell::RefCell;
use std::collections::HashMap;
use ixc_message_api::AccountID;
use ixc_message_api::code::{Code, SystemErrorCode};
use ixc_message_api::handler::{AllocError, Handler, HandlerCode, HostBackend};
use ixc_message_api::packet::MessagePacket;
use ixc_vm_api::{HandlerID, VM};
use ixc_core_macros::message_selector;
use ixc_message_api::header::MessageHeader;

pub struct Hypervisor<ST: StateHandler> {
    vmdata: VMData,
    state_handler: ST,
}

struct VMData {
    vms: HashMap<String, Box<dyn VM>>,
}

impl<ST: StateHandler> Hypervisor<ST> {
    pub fn new(state_handler: ST) -> Self {
        Self {
            vmdata: VMData {
                vms: HashMap::new(),
            },
            state_handler,
        }
    }

    pub fn register_vm(&mut self, name: &str, vm: Box<dyn VM>) {
        self.vmdata.vms.insert(name.to_string(), vm);
    }

    pub fn invoke(&self, message_packet: &mut MessagePacket) -> Code {
        let mut tx = self.state_handler.new_transaction();
        tx.push(message_packet.header().sender_account);
        let mut exec_context = ExecContext {
            vmdata: &self.vmdata,
            tx: RefCell::new(tx),
        };
        let code = exec_context.invoke(message_packet);
        let tx = exec_context.tx.into_inner();
        match code {
            Code::Ok => {
                self.state_handler.commit(tx);
                Code::Ok
            }
            _ => {
                tx.rollback();
                code
            }
        }
    }
}

pub trait StateHandler {
    type Tx: Transaction;
    fn new_transaction(&self) -> Self::Tx;
    fn commit(&self, tx: Self::Tx);
}

pub trait Transaction {
    type KVStore: KVStore;

    fn push(&mut self, account: AccountID);
    fn pop(&mut self, commit: bool);
    fn top(&self) -> Option<AccountID>;
    fn rollback(self);
    fn manager_state(&self) -> &mut Self::KVStore;
}

pub trait KVStore {
    fn get(&self, key: &[u8]) -> Option<Vec<u8>>;
    fn set(&mut self, key: &[u8], value: &[u8]);
    fn delete(&mut self, key: &[u8]);
}

struct ExecContext<'a, TX: Transaction> {
    vmdata: &'a VMData,
    tx: RefCell<TX>,
}

impl<'a, TX: Transaction> ExecContext<'a, TX> {
    fn get_account_handler_id(&self, tx: &mut TX, account_id: AccountID) -> Option<HandlerID> {
        let kv_store = tx.manager_state();
        let key = format!("h:{}", account_id.get());
        let value = kv_store.get(key.as_bytes())?;
        parse_handler_id(&value)
    }
}

fn parse_handler_id(value: &[u8]) -> Option<HandlerID> {
    let mut parts = value.split(|&c| c == b':');
    let vm = parts.next()?;
    let handler_id = parts.next()?;
    Some(HandlerID {
        vm: String::from_utf8(vm.to_vec()).ok()?,
        vm_handler_id: String::from_utf8(handler_id.to_vec()).ok()?,
    })
}

impl<'a, TX: Transaction> HostBackend for ExecContext<'a, TX> {
    fn invoke(&self, message_packet: &mut MessagePacket) -> Code {
        let mut tx = self.tx.try_borrow_mut(); // TODO use try_borrow_mut
        if let Err(_) = tx {
            return Code::SystemError(SystemErrorCode::FatalExecutionError);
        }
        let mut tx = tx.as_mut().unwrap();
        if let Some(account) = tx.top() {
            if message_packet.header().sender_account != account {
                return Code::SystemError(SystemErrorCode::UnauthorizedCallerAccess);
            }
            let target_account = message_packet.header().target_account;
            if target_account.is_null() {
                return self.handle_system_message(&mut tx, message_packet);
            }
            let handler_id = self.get_account_handler_id(&mut tx, target_account);
            if handler_id.is_none() {
                return Code::SystemError(SystemErrorCode::HandlerNotFound);
            }
            let handler_id = handler_id.unwrap();
            let vm = self.vmdata.vms.get(&handler_id.vm);
            if vm.is_none() {
                return Code::SystemError(SystemErrorCode::HandlerNotFound);
            }
            let vm = vm.unwrap();
            tx.push(target_account);
            let code = vm.run_handler(&handler_id.vm_handler_id, message_packet, self);
            match code {
                Code::Ok => {
                    tx.pop(true);
                    Code::Ok
                }
                _ => {
                    tx.pop(false);
                    code
                }
            }
        } else {
            Code::SystemError(SystemErrorCode::FatalExecutionError)
        }
    }

    unsafe fn alloc(&self, layout: Layout) -> Result<*mut u8, AllocError> {
        Ok(std::alloc::alloc(layout))
    }
}

impl<'a, TX: Transaction> ExecContext<'a, TX> {
    fn handle_system_message(&self, tx: &mut TX, message_packet: &mut MessagePacket) -> Code {
        match message_packet.header().message_selector {
            CREATE_SELECTOR => {
                let handler_id = message_packet.in1().get();
                let init_data = message_packet.in2().get();
                // TODO: how do we specify a selector that can only be called by the system?
                let handler_id = parse_handler_id(handler_id);
                if handler_id.is_none() {
                    return Code::SystemError(SystemErrorCode::HandlerNotFound);
                }
                let handler_id = handler_id.unwrap();
                let vm = self.vmdata.vms.get(&handler_id.vm);
                if vm.is_none() {
                    return Code::SystemError(SystemErrorCode::HandlerNotFound);
                }
                let vm = vm.unwrap();

                // TODO create account ID
                // TODO initialize storage
                // TODO push account ID to tx

                let mut on_create_header = MessageHeader::default();
                let on_create_header_ptr: *mut MessageHeader = &mut on_create_header;
                let mut on_create_packet = unsafe { MessagePacket::new(on_create_header_ptr, size_of::<MessageHeader>()) };
                let code = vm.run_handler(&handler_id.vm_handler_id, &mut on_create_packet, self);
                match code {
                    Code::Ok => {
                        tx.pop(true);
                        Code::Ok
                    },
                    _ => {
                        tx.pop(false);
                        code
                    }
                }
            },
            _ => {
                Code::SystemError(SystemErrorCode::HandlerNotFound)
            }
        }
    }
}

const CREATE_SELECTOR: u64 = message_selector!("ixc.account.v1.create");
const ON_CREATE_SELECTOR: u64 = message_selector!("ixc.account.v1.on_create");


#[cfg(test)]
mod tests {
    #[test]
    fn test_parse_handler_id() {
        let value = b"vm1:handler1";
        let handler_id = super::parse_handler_id(value).unwrap();
        assert_eq!(handler_id.vm, "vm1");
        assert_eq!(handler_id.vm_handler_id, "handler1");
    }
}