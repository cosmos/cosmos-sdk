use dashu_int::UBig;
use cosmossdk_core::{Context, Module, Server};
use state_objects::{Map, Pair, UBigMap};
use crate::example::bank::v1::bank::{MsgSend, MsgSendResponse, MsgServer};
use core::borrow::Borrow;
use cosmossdk_macros::{module};

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
    send_enabled: Map<state_objects::Str, bool>,
    balances: UBigMap<Pair<state_objects::Bytes, state_objects::Str>>,
    supplies: UBigMap<state_objects::Str>,
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
        // // checking send enabled uses last block state so no need to synchronize reads
        // if !self.send_enabled.get_last_block(ctx, req.denom.borrow())? {
        //     return ::zeropb::err_msg(Code::Unavailable, "send disabled for denom");
        // }
        //
        // let amount = UBig::from_bytes(req.amount.borrow()).map_err(|_| ::zeropb::err_msg(Code::InvalidArgument, "amount must be a valid UBig"))?;
        //
        // // blocking safe sub must synchronize reads and writes
        // self.balances.safe_sub(ctx, &Pair(req.from.borrow(), req.denom.borrow()), &amount)?;
        //
        // // non-blocking add to recipient won't fail, so no need to synchronize writes
        // self.balances.add_later(ctx, &Pair(req.to.borrow(), req.denom.borrow()), &amount);

        zeropb::ok()
    }
}
