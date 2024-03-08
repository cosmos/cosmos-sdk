use dashu_int::UBig;
use cosmossdk_core::{Code, Router};
use state_objects::{Map, Pair, UBigMap};
use zeropb::Server;
use crate::example::bank::v1::bank::{MsgSend, MsgSendResponse, MsgServer};
use core::borrow::Borrow;

pub mod example {
    pub mod bank {
        pub mod v1 {
            pub mod bank {
                include!("example/bank/v1/bank.rs");
            }
        }
    }
}

pub struct Bank {
    send_enabled: Map<state_objects::Bytes, bool>,
    balances: UBigMap<Pair<state_objects::Bytes, state_objects::Str>>,
    supplies: UBigMap<state_objects::Str>,
}

impl MsgServer for Bank {
    fn send(&self, ctx: &mut ::cosmossdk_core::Context, req: &MsgSend) -> ::zeropb::Result<MsgSendResponse> {
        // checking send enabled uses last block state so no need to synchronize reads
        if !self.send_enabled.get_last_block(ctx, req.from.borrow())? {
            return ::zeropb::err_msg(Code::Unavailable, "send disabled for denom");
        }

        // blocking safe sub must synchronize reads and writes
        self.balances.safe_sub(ctx, &Pair(req.from.borrow(), req.denom.borrow()), &UBig::ZERO)?;

        // non-blocking add to recipient won't fail, so no need to synchronize writes
        self.balances.add(ctx, &Pair(req.to.borrow(), req.denom.borrow()), &UBig::ZERO);

        zeropb::ok()
    }
}