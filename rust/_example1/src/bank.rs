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
    #[map(prefix = 1, key(denom), value(.))]
    denom_owners: Map<Denom, DenomOwnerInfo>,

    #[map(prefix = 2, key(addess, denom), value(balance))]
    balances: Map<(Address, Denom), u128>,

    #[map(prefix = 3, key(denom), value(supply))]
    supply: Map<Denom, u128>,
}

// a trait that accounts can implement to intercept send calls
#[service]
pub trait DenomOnSend {
    fn on_send(&self, ctx: &Context, from_address: &Address, to_address: &Address, coin: &Coin) -> Result<()>;
}

#[service]
pub trait DenomOverride {
    fn balance(&self, ctx: &Context, address: &Address, denom: &Denom) -> Result<u128>;
    fn supply(&self, ctx: &Context, denom: &Denom) -> Result<u128>;
}

#[derive(Serializable)]
pub struct DenomOwnerInfo {
    owner: Address,

    // if true, then the owner has overridden balance, supply and send tracking
    // this value can't be changed from its initial value (that would require a major data migration)
    has_override: bool,
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

#[service(proto_package = "cosmos.bank.v2")]
pub trait BankV2 {
    fn create_denom(&self, ctx: &mut Context, denom: &Denom, owner: &Address, has_override: bool) -> Result<()>;
    fn send(&self, ctx: &mut Context, to_address: &Address, amount: &[Coin]) -> Result<()>;
    fn mint(&self, ctx: &mut Context, to_address: &Address, coin: &Coin) -> Result<()>;
    fn burn(&self, ctx: &mut Context, from_address: &Address, coin: &Coin) -> Result<()>;
    fn balance(&self, ctx: &Context, address: &Address, denom: &Denom) -> Result<u128>;
    fn supply(&self, ctx: &Context, denom: &Denom) -> Result<u128>;
}

impl BankV2 for Bank {
    fn create_denom(&self, ctx: &mut Context, denom: &Denom, owner: &Address, has_override: bool) -> Result<()> {
        if self.denom_owners.get(ctx, denom)?.is_some() {
            return Err("denom already exists".to_string());
        }
        self.denom_owners.set(ctx, denom, &DenomOwnerInfo {
            owner: owner.clone(),
            has_override,
        })
    }

    fn send(&self, ctx: &mut Context, to_address: &Address, amount: &[Coin]) -> Result<()> {
        self::BankMsg::send(self, ctx, &ctx.self_address(), to_address, amount)
    }

    fn balance(&self, ctx: &Context, address: &Address, denom: &Denom) -> Result<u128> {
        self::BankQuery::balance(self, ctx, address, denom)
    }

    fn mint(&self, ctx: &mut Context, to_address: &Address, coin: &Coin) -> Result<()> {
        if let Some(owner_info) = self.denom_owners.get(ctx, &coin.denom)? {
            if owner_info.has_override {
                return Err("denom has balance and supply tracking override".to_string());
            }

            let supply = self.supply.get(ctx, &coin.denom)?.unwrap_or(0);
            self.supply.set(ctx, &coin.denom, &(supply + coin.amount))?;
            self::BankMsg::send(self, ctx, &ctx.self_address(), to_address, &[coin.clone()])
        } else {
            Err("denom not found".to_string())
        }
    }

    fn burn(&self, ctx: &mut Context, from_address: &Address, coin: &Coin) -> Result<()> {
        if let Some(owner_info) = self.denom_owners.get(ctx, &coin.denom)? {
            if owner_info.has_override {
                return Err("denom has balance and supply tracking override".to_string());
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

    fn supply(&self, ctx: &Context, denom: &Denom) -> Result<u128> {
        if let Some(owner_info) = self.denom_owners.get(ctx, denom)? {
            if owner_info.has_override {
                let supply_client = DenomOverrideClient(owner_info.owner);
                return supply_client.supply(ctx, denom);
            }
        }

        self.supply.get(ctx, denom).map(|supply| supply.unwrap_or(0))
    }
}

impl BankMsg for Bank {
    fn send(&self, ctx: &mut Context, from_address: &Address, to_address: &Address, amount: &[Coin]) -> Result<()> {
        for coin in amount {
            if let Some(owner_info) = self.denom_owners.get(ctx, &coin.denom)? {
                let can_send_client = DenomOnSendClient(owner_info.owner);
                if owner_info.has_override {
                    // if the owner has an override, then we just call on_send to process the transfer
                    // if it's not implemented then this call will error, which is what we want
                    can_send_client.on_send(ctx, from_address, to_address, coin)?;
                    continue;
                }

                // we dynamically check if there is an on_send method implemented,
                // as an upgradeable account may dynamically add or remove the on_send method
                if can_send_client.on_send_implemented(ctx)? {
                    // we'll return with an error here if the on_send method returns an error blocking the send
                    can_send_client.on_send(ctx, from_address, to_address, coin)?
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
        if let Some(owner_info) = self.denom_owners.get(ctx, denom)? {
            if owner_info.has_override {
                let balance_client = DenomOverrideClient(owner_info.owner);
                return balance_client.balance(ctx, address, denom);
            }
        }

        self.balances.get(ctx, &(address.clone(), denom.clone())).map(|balance| balance.unwrap_or(0))
    }
}
