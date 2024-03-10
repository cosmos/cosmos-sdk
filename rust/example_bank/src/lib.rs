use dashu_int::UBig;
use cosmossdk_core::{Code, Context, Module, Router};
use state_objects::{Index, Map, Str, UBigMap};
use crate::example::bank::v1::bank::{MsgSend, MsgSendResponse, MsgServer};
use core::borrow::Borrow;
use cosmossdk_macros::{module};
use state_objects_macros::State;

pub mod example {
    pub mod bank {
        pub mod v1 {
            pub mod bank {
                include!("example/bank/v1/bank.rs");
            }
        }
    }
}

#[module(name = "example.bank.v1", services=[MsgServer])]
pub struct Bank {
    state: BankState,
}

#[derive(State)]
pub struct BankState {
    #[map(prefix=1, key(denom), value(enabled))]
    send_enabled: Map<String, bool>,

    #[map(prefix=2, key(address, denom), value(balance))]
    balances: UBigMap<(Vec<u8>, String)>,

    #[map(prefix=3, key(module, denom), value(balance))]
    module_balances: UBigMap<(String, String)>,

    #[map(prefix=4, key(denom), value(supply))]
    supplies: UBigMap<String>,

    #[index(prefix=5, on(balances(denom, address)))]
    balances_by_denom: Index<(String, Vec<u8>), UBig>,

    #[index(prefix=6, on(balances(denom, module)))]
    module_balances_by_denom: Index<(String, String), UBig>,
}

// impl Module for Bank {
//     fn route(&self, route_id: u64, ctx: &mut Context, req: *mut u8, res: *mut *mut u8) -> Code {
//         // service id is second to last byte of route id
//         let service_id = (route_id >> 8) & 0xffu64;
//         // method id is last byte of route id
//         let method_id = route_id & 0xffu64;
//         match service_id {
//             0x0 => <dyn MsgServer as Server>::route(self, method_id, ctx, req, res),
//             _ => Code::Unimplemented,
//         }
//     }
// }

impl MsgServer for Bank {
    fn send(&self, ctx: &mut Context, req: &MsgSend) -> ::zeropb::Result<MsgSendResponse> {
        // checking send enabled uses last block state so no need to synchronize reads
        if !self.state.send_enabled.get_stale(ctx, req.denom.borrow())? {
            return ::zeropb::err_msg(Code::Unavailable, "send disabled for denom");
        }

        let amount = UBig::from_bytes(req.amount.borrow()).map_err(|_| ::zeropb::err_msg(Code::InvalidArgument, "amount must be a valid UBig"))?;

        // blocking safe sub must synchronize reads and writes
        self.state.balances.safe_sub(ctx, &(req.from.borrow(), req.denom.borrow()), &amount)?;
        //
        // non-blocking add to recipient won't fail, so no need to synchronize writes
        self.state.balances.add_lazy(ctx, &(req.to.borrow(), req.denom.borrow()), &amount);

        zeropb::ok()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmossdk_core::testing::MockApp;
    use cosmossdk_core::Server;
    use cosmossdk_core::store::{MockStore, Store};

    #[test]
    fn test_send() {
        let mut app = MockApp::new();
        let mut mock_store = MockStore::new();
        app.add_module("bank", Bank::new(), ());
        app.add_mock_server(mock_store as &dyn Store);
    }
}