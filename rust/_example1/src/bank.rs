#![derive_module(Bank)]

use arrayvec::ArrayString;
use cosmos_core::{Address, Context, Map, Result};
use cosmos_core_macros::{service, Serializable, proto_method, derive_module, State};

type Denom = ArrayString<256>;

#[derive(Serializable, Clone)]
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

pub trait BankV2 {
    fn create_denom(&self, ctx: &mut Context, denom: &Denom, owner: &Address) -> Result<()>;
    fn send(&self, ctx: &mut Context, to_address: &Address, amount: &[Coin]) -> Result<()>;
    fn balance(&self, ctx: &Context, address: &Address, denom: &Denom) -> Result<u128>;
    fn mint(&self, ctx: &mut Context, to_address: &Address, coin: &Coin) -> Result<()>;
    fn burn(&self, ctx: &mut Context, from_address: &Address, coin: &Coin) -> Result<()>;
}

#[service]
pub trait DenomCanSend {
    fn can_send(&self, ctx: &Context, from_address: &Address, to_address: &Address, coin: &Coin) -> Result<bool>;
}

#[service]
pub trait DenomOverride {
    fn send(&self, ctx: &Context, from_address: &Address, to_address: &Address, coin: &Coin) -> Result<bool>;
    fn balance(&self, ctx: &Context, address: &Address, denom: &Denom) -> Result<u128>;
    fn mint(&self, ctx: &mut Context, to_address: &Address, coin: &Coin) -> Result<()>;
    fn burn(&self, ctx: &mut Context, from_address: &Address, coin: &Coin) -> Result<()>;
    fn supply(&self, ctx: &Context, denom: &Denom) -> Result<u128>;
}

impl BankV2 for Bank {
    fn create_denom(&self, ctx: &mut Context, denom: &Denom, owner: &Address) -> Result<()> {
        if self.denom_owners.get(ctx, denom)?.is_some() {
            return Err("denom already exists".to_string());
        }
        self.denom_owners.set(ctx, denom, owner)
    }

    fn send(&self, ctx: &mut Context, to_address: &Address, amount: &[Coin]) -> Result<()> {
        self::BankMsg::send(self, ctx, &ctx.self_address(), to_address, amount)
    }

    fn balance(&self, ctx: &Context, address: &Address, denom: &Denom) -> Result<u128> {
        self::BankQuery::balance(self, ctx, address, denom)
    }

    fn mint(&self, ctx: &mut Context, to_address: &Address, coin: &Coin) -> Result<()> {
        if let Some(denom_owner) = self.denom_owners.get(ctx, &coin.denom)? {
            let mint_client = DenomOverrideClient(denom_owner);
            if mint_client.mint_implemented(ctx)? {
                // if mint is implemented, then we can mint using that method and not do any other logic
                return mint_client.mint(ctx, to_address, coin);
            }

            let supply = self.supply.get(ctx, &coin.denom)?.unwrap_or(0);
            self.supply.set(ctx, &coin.denom, &(supply + coin.amount))?;
            self::BankMsg::send(self, ctx, &ctx.self_address(), to_address, &[coin.clone()])
        } else {
            Err("denom not found".to_string())
        }
    }

    fn burn(&self, ctx: &mut Context, from_address: &Address, coin: &Coin) -> Result<()> {
        if let Some(denom_owner) = self.denom_owners.get(ctx, &coin.denom)? {
            let burn_client = DenomOverrideClient(denom_owner);
            if burn_client.burn_implemented(ctx)? {
                // if burn is implemented, then we can burn using that method and not do any other logic
                return burn_client.burn(ctx, from_address, coin);
            }

            let from_balance = self.balances.get(ctx, &(from_address.clone(), coin.denom.clone()))?.unwrap_or(0);
            if from_balance < coin.amount {
                return Err("insufficient funds".to_string());
            }
            self.balances.set(ctx, &(from_address.clone(), coin.denom.clone()), &(from_balance - coin.amount))?;

            let supply = self.supply.get(ctx, &coin.denom)?.unwrap_or(0);
            if supply < coin.amount {
                return Err("insufficient supply".to_string());
            }
            self.supply.set(ctx, &coin.denom, &(supply - coin.amount))
        } else {
            Err("denom not found".to_string())
        }
    }
}

impl BankMsg for Bank {
    fn send(&self, ctx: &mut Context, from_address: &Address, to_address: &Address, amount: &[Coin]) -> Result<()> {
        for coin in amount {
            if let Some(denom_owner) = self.denom_owners.get(ctx, &coin.denom)? {
                let send_client = DenomOverrideClient(denom_owner);
                if send_client.send_implemented(ctx)? {
                    // if send is implemented, then we can send using that method and not do any other logic
                    send_client.send(ctx, from_address, to_address, coin)?;
                    continue;
                }

                let can_send_client = DenomCanSendClient(denom_owner);
                if can_send_client.can_send_implemented(ctx)? {
                    if !can_send_client.can_send(ctx, from_address, to_address, coin)? {
                        return Err("send blocked".to_string());
                    }
                }
            } else {
                return Err("denom not found".to_string());
            }

            let from_balance = self.balances.get(ctx, &(from_address.clone(), coin.denom.clone()))?.unwrap_or(0);
            if from_balance < coin.amount {
                return Err("insufficient funds".to_string());
            }
            let to_balance = self.balances.get(ctx, &(to_address.clone(), coin.denom.clone()))?.unwrap_or(0);
            self.balances.set(ctx, &(from_address.clone(), coin.denom.clone()), &(from_balance + coin.amount))?;
            self.balances.set(ctx, &(to_address.clone(), coin.denom.clone()), &(to_balance + coin.amount))?;
        }
        Ok(())
    }
}

impl BankQuery for Bank {
    fn balance(&self, ctx: &Context, address: &Address, denom: &Denom) -> Result<u128> {
        if let Some(denom_owner) = self.denom_owners.get(ctx, denom)? {
            let balance_client = DenomOverrideClient(denom_owner);
            if balance_client.balance_implemented(ctx)? {
                // if balance is implemented, then we can get the balance using that method and not do any other logic
                return balance_client.balance(ctx, address, denom);
            }
        }

        self.balances.get(ctx, &(address.clone(), denom.clone())).map(|balance| balance.unwrap_or(0))
    }
}