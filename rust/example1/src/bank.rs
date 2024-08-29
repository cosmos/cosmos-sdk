use arrayvec::ArrayString;
use crypto_bigint::U256;
use cosmos_core::{Address, Context, Result};
use cosmos_core_macros::{service, Serializable, proto_method};

type Denom = ArrayString<256>;

#[derive(Serializable)]
#[proto_name="cosmos.bank.v1beta1.Coin"]
pub struct Coin {
    #[proto(tag="1")]
    denom: Denom,

    #[proto(tag="2", type="string")]
    amount: U256,
}

#[service(proto_package="cosmos.bank.v1beta1")]
pub trait BankMsg {
    #[proto_method(name="MsgSend", v1_signer="from_address")]
    fn send(&self, ctx: &mut Context, from_address: &Address, to_address: &Address, amount: &[Coin]) -> Result<()>;
}