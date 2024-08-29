use proc_macro::TokenStream;
use quote::{format_ident, quote};
use syn::{parse_macro_input, ItemTrait, TraitItem};

#[proc_macro_attribute]
pub fn service(_: TokenStream, item: TokenStream) -> TokenStream {
    let item2: proc_macro2::TokenStream = item.clone().into();
    let input = parse_macro_input!(item as ItemTrait);
    let trait_name = &input.ident;
    let client_struct_name = format_ident!("{}Client", trait_name);

    let methods = input.items.iter().filter_map(|ti| {
        match ti {
            TraitItem::Fn(f) => {
                let name = &f.sig.ident;
                let args = &f.sig.inputs;
                let ret = &f.sig.output;
                Some(quote! {
                    fn #name(#args) #ret {
                        todo!()
                    }
                })
            }
            _ => None
        }
    });
    let tokens = quote! {
        #item2

        pub struct #client_struct_name {}

        impl #trait_name for #client_struct_name {
            #(#methods)*
        }
    };

    tokens.into()
}

#[proc_macro_attribute]
pub fn proto_method(_: TokenStream, item: TokenStream) -> TokenStream {
    item
}

#[proc_macro_derive(Account)]
pub fn derive_account(item: TokenStream) -> TokenStream {
    TokenStream::new()
}

#[proc_macro_derive(Module)]
pub fn derive_module(item: TokenStream) -> TokenStream {
    TokenStream::new()
}

#[proc_macro_derive(State, attributes(map, item, table))]
pub fn derive_state(item: TokenStream) -> TokenStream {
    TokenStream::new()
}

#[proc_macro_derive(Serializable, attributes(proto))]
pub fn derive_serializable(item: TokenStream) -> TokenStream {
    TokenStream::new()
}

#[proc_macro_derive(Table)]
pub fn derive_table(item: TokenStream) -> TokenStream {
    TokenStream::new()
}