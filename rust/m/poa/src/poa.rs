#![allow(missing_docs)]
#[ixc::handler(POA)]
pub mod poa {
    use mockall::automock;
    use ixc::*;
    use ixc_core::error::unimplemented_ok;
    use ixc_core::handler::Service;


    #[derive(Resources)]
    pub struct POA {
        #[state(prefix = 1, key(address), value(validator))]
        pub(crate) validators: Map<AccountID, Validator>,
        #[state(prefix = 2)]
        pub(crate) validator_set: Item<ValidatorSet>,
        #[state(prefix = 3)]
        pub(crate) total_power: Item<u128>, 
        #[state(prefix = 4)]
        admin: Item<AccountID>,
    }

    #[derive(SchemaValue)]
    #[non_exhaustive]
    pub struct Validator {
        pub address: AccountID,
        pub consensus_pubkey: Vec<u8>,
        pub power: u128,
    }

    #[derive(SchemaValue)]
    #[non_exhaustive]
    pub struct ValidatorSet {
        pub validators: Vec<Validator>,
        pub total_power: u128,
    }

    #[handler_api]
    pub trait POAAPI {
        fn add_validator(&self, ctx: &mut Context, validator: Validator) -> Result<()>;
        fn remove_validator(&self, ctx: &mut Context, validator: AccountID) -> Result<()>;
        fn update_validator_power(&self, ctx: &mut Context, validator: AccountID, power: u128) -> Result<()>;
        fn get_validator(&self, ctx: &Context, validator: AccountID) -> Result<Validator>;
        fn get_validator_set(&self, ctx: &Context) -> Result<ValidatorSet>;
    }

   #[derive(SchemaValue)]
    #[non_exhaustive]
    pub struct EventAddValidator {
        pub validator: Validator,
    }

  #[derive(SchemaValue)]
    #[non_exhaustive]
    pub struct EventRemoveValidator {
        pub validator: AccountID,
    }

    #[derive(SchemaValue)]
    #[non_exhaustive]
    pub struct EventUpdateValidatorPower {
        pub validator: AccountID,
        pub power: u128,
    }

    impl POA {
        #[on_create]
        pub fn create(&self, ctx: &mut Context) -> Result<()> {
            self.admin.set(ctx, ctx.caller())?;
            Ok(())
        }
    }

    #[publish]
    impl POAAPI for POA {
        fn add_validator(&self, ctx: &mut Context, validator: Validator) -> Result<()> {
            let admin = self.admin.get(ctx)?;
            ensure!(admin == ctx.caller(), "not authorized");
            self.validators.set(ctx, validator.address, validator)?;
            self.validator_set.set(ctx, ValidatorSet {
                validators: vec![validator.clone()],
                total_power: validator.power,
            })?;
            // increase total power
            self.total_power.set(ctx, self.total_power.get(ctx)? + validator.power)?;
            Ok(())
        }

        fn remove_validator(&self, ctx: &mut Context, validator: AccountID) -> Result<()> {
            let admin = self.admin.get(ctx)?;
            ensure!(admin == ctx.caller(), "not authorized");
            self.validators.remove(ctx, validator)?;
            self.validator_set.set(ctx, ValidatorSet {
                validators: self.validator_set.get(ctx)?.validators.into_iter().filter(|v| v.address != validator).collect(),
                total_power: self.validator_set.get(ctx)?.total_power - self.validators.get(ctx, validator)?.power,
            })?;
            //decrease total power
            self.total_power.set(ctx, self.total_power.get(ctx)? - self.validators.get(ctx, validator)?.power)?;
            Ok(())
        }

        fn update_validator_power(&self, ctx: &mut Context, validator: AccountID, power: u128) -> Result<()> {
            let admin = self.admin.get(ctx)?;
            ensure!(admin == ctx.caller(), "not authorized");
            let mut validator = self.validators.get(ctx, validator)?;
            // remove the previous validator power from total power and add the new validator power to total power
            self.total_power.set(ctx, self.total_power.get(ctx)? - validator.power + power)?;

            validator.power = power;
            self.validators.set(ctx, validator.address, validator)?;
            self.validator_set.set(ctx, ValidatorSet {
                validators: self.validator_set.get(ctx)?.validators.into_iter().map(|v| if v.address == validator { validator.clone() } else { v }).collect(),
                total_power: self.validator_set.get(ctx)?.total_power - validator.power + power,
            })?;
            Ok(())
        }

        fn get_validator(&self, ctx: &Context, validator: AccountID) -> Result<Validator> {
            self.validators.get(ctx, validator)
        }

        fn get_validator_set(&self, ctx: &Context) -> Result<ValidatorSet> {
            self.validator_set.get(ctx)
        }
    }
}
