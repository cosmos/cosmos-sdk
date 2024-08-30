#![derive_module(Bank)]

use arrayvec::ArrayString;
use cosmos_core::{Address, Context, Map, Result};
use cosmos_core_macros::{service, Serializable, proto_method, derive_module, State};

type Denom = ArrayString<256>;

#[derive(Serializable)]
#[proto(name = "cosmos.bank.v1beta1.Coin")]
pub struct Coin {
    #[proto(tag = "1")]
    denom: Denom,

    #[proto(tag="2", type="string")]
    amount: u128,
}


#[derive(State)]
pub struct Bank {
    #[map(prefix = 1, key(denom), value(owner))]
    denom_owners: Map<Denom, Address>,

    #[map(prefix = 2, key(addess, denom), value(balance))]
    balances: Map<(Address, Denom), u128>,

    #[map(prefix = 3, key(denom), value(supply))]
    supply: Map<Denom, u128>,
}

#[service(proto_package = "cosmos.bank.v1beta1")]
pub trait BankMsg {
    #[proto_method(name = "MsgSend", v1_signer = "from_address")]
    fn send(&self, ctx: &mut Context, from_address: &Address, to_address: &Address, amount: &[Coin]) -> Result<()>;
}

#[service(proto_package = "cosmos.bank.v1beta1")]
pub trait BankQuery {
    #[proto_method(name = "QueryBalance")]
    fn balance(&self, ctx: &Context, address: &Address, denom: &Denom) -> Result<u128>;
}

#[service]
pub trait DenomBalance {
    fn send(&self, ctx: &Context, from_address: &Address, to_address: &Address, coin: &Coin) -> Result<bool>;
    fn balance(&self, ctx: &Context, address: &Address, denom: &Denom) -> Result<u128>;
}

#[service]
pub trait DenomCanSend {
    fn can_send(&self, ctx: &Context, from_address: &Address, to_address: &Address, coin: &Coin) -> Result<bool>;
}

impl BankMsg for Bank {
    fn send(&self, ctx: &mut Context, from_address: &Address, to_address: &Address, amount: &[Coin]) -> Result<()> {
        for coin in amount {
            let send_client = DenomBalanceClient(self.denom_owners.get(ctx, &coin.denom)?);
            if send_client.send_implemented(ctx)? {
                // if send is implemented, then we can send using that method and not do any other logic
                send_client.send(ctx, from_address, to_address, coin)?;
                continue;
            }

            let can_send_client = DenomCanSendClient(self.denom_owners.get(ctx, &coin.denom)?);
            if can_send_client.can_send_implemented(ctx)? {
                if !can_send_client.can_send(ctx, from_address, to_address, coin)? {
                    return Err("send blocked".to_string());
                }
            }

            let from_balance = self.balances.get(ctx, &(from_address.clone(), coin.denom.clone()))?;
            if from_balance < coin.amount {
                return Err("insufficient funds".to_string());
            }
            let to_balance = self.balances.get(ctx, &(to_address.clone(), coin.denom.clone()))?;
            self.balances.set(ctx, &(from_address.clone(), coin.denom.clone()), &(from_balance + coin.amount))?;
            self.balances.set(ctx, &(to_address.clone(), coin.denom.clone()), &(to_balance + coin.amount))?;
        }
        Ok(())
    }
}

impl BankQuery for Bank {
    fn balance(&self, ctx: &Context, address: &Address, denom: &Denom) -> Result<u128> {
        let balance_client = DenomBalanceClient(self.denom_owners.get(ctx, denom)?);
        if balance_client.balance_implemented(ctx)? {
            // if balance is implemented, then we can get the balance using that method and not do any other logic
            return balance_client.balance(ctx, address, denom);
        }

        self.balances.get(ctx, &(address.clone(), denom.clone()))
    }
}