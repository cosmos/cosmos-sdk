use proc_macro::TokenStream;
use syn::parse_macro_input;

#[proc_macro_derive(ZeroCopy)]
pub fn derive_zero_copy(input: TokenStream) -> TokenStream {
    // let input = parse_macro_input!(input as DeriveInput);
    todo!()
}
