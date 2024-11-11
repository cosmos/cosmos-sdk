#![allow(missing_docs)]
#[ixc::handler(ModuleManager)]
mod module_manager {
    use ixc::*;
    use ixc_core::message::Message;
    use ixc_core::result::ClientResult;
    use ixc_schema::value::OptionalValue;

    #[derive(Resources)]
    pub struct ModuleManager {
        #[state(prefix = 1, key(modules), value(account))]
        modules: Map<Str, AccountID>,

        #[state(prefix = 2, key(message_name), value(account))]
        message_handlers: Map<Str, AccountID>,
    }

    pub trait ModuleRouter {
        fn invoke<'a, 'b, M: Message<'b>>(context: &'a Context, message: M) -> ClientResult<<M::Response<'a> as OptionalValue<'a>>::Value, M::Error>;
    }
}

fn main() {}