#![allow(missing_docs)]
#[ixc::handler(Bank)]
pub mod bank {
    use ixc::*;
    use ixc_core::handler::ClientFactory;

    #[derive(Resources)]
    pub struct Bank {
        #[state(key(address, denom), value(amount))]
        balances: AccumulatorMap<(AccountID, Str)>,
    }

    #[derive(SchemaValue)]
    #[sealed]
    pub struct Coin<'a> {
        pub denom: &'a str,
        pub amount: u128,
    }

    #[handler_api]
    pub trait BankAPI {
        fn get_balance(&self, ctx: &Context, account: AccountID, denom: &str) -> Result<u128>;
        fn send<'a>(&self, ctx: &'a mut Context, to: AccountID, amount: &[Coin<'a>], evt: EventBus<EventSend<'_>>) -> Result<()>;
    }

    #[derive(SchemaValue)]
    #[non_exhaustive]
    pub struct EventSend<'a> {
        pub from: AccountID,
        pub to: AccountID,
        pub coin: Coin<'a>,
    }

    impl Bank {
        #[on_create]
        fn create(&self, ctx: &mut Context) -> Result<()> {
            Ok(())
        }
    }

    impl BankAPI for Bank {
        fn get_balance(&self, ctx: &Context, account: AccountID, denom: &str) -> Result<u128> {
            self.balances.get(ctx, (account, denom))
        }

        fn send(&self, ctx: &mut Context, to: AccountID, amount: &[Coin], evt: EventBus<EventSend>) -> Result<()> {
            for coin in amount {
                self.balances.safe_sub(ctx, (ctx.caller(), coin.denom), coin.amount)?;
                self.balances.add(ctx, (to, coin.denom), coin.amount)?;
                // evt.emit(EventSend {
                //     from: ctx.sender(),
                //     to,
                //     coin: coin.clone(),
                // })?;
            }
            Ok(())
        }
    }
    // unsafe impl ::ixc_core::routes::Router for dyn BankAPI {
    //     const SORTED_ROUTES: &'static [::ixc_core::routes::Route<Self>] = &::ixc_core::routes::sort_routes([(<BankAPIGetBalance as ::ixc_core::message::Message>::SELECTOR, |h: &dyn BankAPI, packet, cb, a| {
    //         unsafe {
    //             let cdc = <BankAPIGetBalance as ::ixc_core::message::Message>::Codec::default();
    //             let in1 = packet.header().in_pointer1.get(packet);
    //             let mut ctx = ::ixc_core::Context::new(packet.header().context_info, cb);
    //             let BankAPIGetBalance {
    //                 account, denom, } = ::ixc_schema::codec::decode_value::<BankAPIGetBalance>(&cdc, in1, ctx.memory_manager()).map_err(|e| ::ixc_message_api::handler::HandlerError::Custom(0))?;
    //             let res = h.get_balance(core::mem::transmute(&ctx), account, denom).map_err(|e| ::ixc_message_api::handler::HandlerError::Custom(0))?;
    //             ::ixc_core::low_level::encode_optional_to_out1::<<BankAPIGetBalance as ::ixc_core::message::Message<'_>>::Response<'_>>(&cdc, &res, a, packet).map_err(|e| ::ixc_message_api::handler::HandlerError::Custom(0))
    //         }
    //     }), (<BankAPISend as ::ixc_core::message::Message>::SELECTOR, |h: &dyn BankAPI, packet, cb, a| {
    //         unsafe {
    //             let cdc = <BankAPISend as ::ixc_core::message::Message>::Codec::default();
    //             let in1 = packet.header().in_pointer1.get(packet);
    //             let mut ctx = ::ixc_core::Context::new(packet.header().context_info, cb);
    //             let BankAPISend {
    //                 to, amount, } = ::ixc_schema::codec::decode_value::<BankAPISend>(&cdc, in1, ctx.memory_manager()).map_err(|e| ::ixc_message_api::handler::HandlerError::Custom(0))?;
    //             let res = h.send(core::mem::transmute(&ctx), to, amount, Default::default()).map_err(|e| ::ixc_message_api::handler::HandlerError::Custom(0))?;
    //             ::ixc_core::low_level::encode_optional_to_out1::<<BankAPISend as ::ixc_core::message::Message<'_>>::Response<'_>>(&cdc, &res, a, packet).map_err(|e| ::ixc_message_api::handler::HandlerError::Custom(0))
    //         }
    //     }), ]);
    // }
}

#[cfg(test)]
mod tests {
    use super::bank::*;
    use ixc_testing::*;

    #[test]
    fn test() {
        let mut app = TestApp::default();
        app.register_handler::<Bank>().unwrap();
    }
}

fn main() {}