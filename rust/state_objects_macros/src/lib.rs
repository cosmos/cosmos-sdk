use proc_macro::TokenStream;

#[proc_macro_derive(State, attributes(map, index, unique_index))]
pub fn state(input: TokenStream) -> TokenStream {
    TokenStream::new()
}
