use cosmos_core::Address;
use cosmos_core::sync::{Map, PrepareContext, AsyncResponse};

pub trait BankSend {
    fn send(&self, ctx: PrepareContext, to_address: &Address, denom: &str) -> AsyncResponse<u128>;
}

pub struct Bank {
    balances: Map<(Address, String), u128>,
}

pub struct EventSend {
    from: Address,
    to: Address,
    denom: String,
    amount: u128,
}

impl BankSend for Bank {
    fn send(&self, ctx: PrepareContext<EventSend>, to_address: &Address, denom: &str) -> AsyncResponse<u128> {
        let get_from = self.balances.prepare_get(ctx.new(), &(ctx.caller().clone(), denom.to_string()))?;
        let get_to = self.balances.prepare_get(ctx.new(), &(to_address.clone(), denom.to_string()))?;
        let set_from = self.balances.prepare_set(ctx.new(), &(ctx.caller().clone(), denom.to_string()))?;
        let set_to = self.balances.prepare_set(ctx.new(), &(to_address.clone(), denom.to_string()))?;
        ctx.exec(|ctx, amount: u128| {
            let from = get_from.exec(ctx.new(), ())?.read()?;
            if from < &amount {
                return Err("insufficient funds".to_string());
            }
            let to = get_to.exec(ctx.new(), ())?.read()?;
            set_from.exec(ctx.new(), from - amount)?;
            set_to.exec(ctx.new(), to + amount)?;
            ctx.emit(EventSend{
                from: ctx.self_address().clone(),
                to: to_address.clone(),
                denom: denom.into(),
                amount,
            })?;
            ctx.ok(())
        })
    }
}
