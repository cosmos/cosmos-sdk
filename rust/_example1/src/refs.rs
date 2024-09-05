mod payment_voucher {
    #![derive_object(PaymentVoucher)]

    use cosmos_core::Context;
    use crate::bank;
    use crate::refs::builtin;

    pub struct PaymentVoucher {
        pub amount: bank::Coin,
        capability: builtin::ServiceCapability<dyn bank::BankV2>,
    }

    impl PaymentVoucher {
        pub fn new(ctx: &Context, amount: bank::Coin) -> Self {
            Self { amount, capability: builtin::ServiceCapability::new(ctx) }
        }

        pub fn cash_in(self, ctx: &mut Context) -> cosmos_core::Result<()> {
            self.capability.get().send(ctx, &ctx.self_address(), &[self.amount])
        }
    }
}

mod digital_good {
    #![derive_module(DigitalGood)]

    use cosmos_core::{Context, Item, Map};
    use cosmos_core_macros::{service, State};
    use crate::bank::Denom;
    use crate::refs::payment_voucher;

    #[derive(State)]
    pub struct DigitalGood {
        #[map(prefix=1, key(owner))]
        purchasers: Map<cosmos_core::Address, ()>,

        #[item(prefix=2)]
        price: Item<u128>,

        #[item(prefix=3)]
        buy_denom: Item<Denom>,
    }

    #[service]
    pub trait BuyIt {
        fn buy_it(&self, ctx: &mut Context, voucher: &payment_voucher::PaymentVoucher::Ref) -> cosmos_core::Result<()>;
    }

    impl BuyIt for DigitalGood {
        fn buy_it(&self, ctx: &mut Context, voucher: &payment_voucher::PaymentVoucher::Ref) -> cosmos_core::Result<()> {
            if self.purchasers.get(ctx, &ctx.sender())?.is_some() {
                return Err("already purchased".to_string());
            }

            if self.price.get(ctx)? != voucher.amount.amount && self.buy_denom.get(ctx)? != voucher.amount.denom {
                return Err("wrong price".to_string());
            }

            voucher.cash_in(ctx)?;

            self.purchasers.set(ctx, &ctx.sender(), &())
        }
    }
}

mod alice {
    #![derive_account(Alice)]

    use cosmos_core::{Address, Context};
    use crate::bank::Coin;
    use crate::refs::digital_good;
    use crate::refs::digital_good::BuyIt;
    use crate::refs::payment_voucher::PaymentVoucher;

    pub struct Alice {}

    impl Alice {
        fn buy_the_stuff(&self, ctx: &mut Context) -> cosmos_core::Result<()> {
            let payment_voucher = PaymentVoucher::new(ctx, Coin { amount: 100, denom: "foobar".into() });
            let buy_it_client = digital_good::BuyItClient(Address::default());
            buy_it_client.buy_it(ctx, payment_voucher)
        }
    }
}

pub mod builtin {
    #![derive_object(ServiceCapability)]

    use cosmos_core::{Address, Context, Service};

    // This is a built-in struct that allows someone to create an unforgeable capability to call a service on another account's behalf.
    // It could potentially be stored, sent, or shared with others.
    pub struct ServiceCapability<T> {
        _phantom: core::marker::PhantomData<T>,
        granter: Address,
    }

    impl<T: Service> ServiceCapability<T> {
        pub(crate) fn new(ctx: &Context) -> Self {
            Self {
                _phantom: core::marker::PhantomData,
                granter: ctx.sender(),
            }
        }

        pub fn get(&self) -> Box<T> {
            T::client_with_ctx(|ctx| {
                ctx.set_sender(self.granter);
            })
        }
    }
}

