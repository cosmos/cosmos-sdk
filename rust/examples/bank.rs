#[interchain_prelude::module_handler(Bank)]
pub mod bank {
    use interchain_prelude::*;

    pub struct Bank {
        balances: Map<(Address, String), u128>,
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
            todo!()
        }
    }
}

fn main() {}