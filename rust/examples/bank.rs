#![allow(missing_docs)]
#[interchain_sdk::module_handler(Bank)]
pub mod bank {
    use interchain_sdk::*;

    pub struct Bank {
        #[schema(name(address, denom), value(amount))]
        balances: UIntMap<(Address, String), u128>,
    }

    #[derive(StructCodec)]
    pub struct Coin<'a> {
        pub denom: &'a str,
        pub amount: u128,
    }

    #[module_api]
    pub trait BankAPI {
        fn send(&self, ctx: &mut Context, to: Address, amount: &[Coin], evt: &mut EventBus<EventSend>) -> Response<()>;
    }

    #[derive(StructCodec)]
    pub struct EventSend<'a> {
        pub from: Address,
        pub to: Address,
        pub coin: Coin<'a>,
    }

    impl BankAPI for Bank {
        fn send(&self, ctx: &mut Context, to: Address, amount: &[Coin], evt: &mut EventBus<EventSend>) -> Response<()> {
            for coin in amount {
                self.balances.safe_sub(ctx, (ctx.sender(), coin.denom), coin.amount)?;
                self.balances.add(ctx, (to, coin.denom), coin.amount)?;
                evt.emit(EventSend {
                    from: ctx.sender(),
                    to,
                    coin: coin.clone(),
                })?;
            }
            Ok(())
        }
    }
}

fn main() {}