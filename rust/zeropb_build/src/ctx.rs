use crate::opts::Options;

#[derive(Default)]
pub(crate) struct Context {
    pub(crate) opts: Options,
    pub(crate) header_items: Vec<syn::Item>,
    pub(crate) body_items: Vec<syn::Item>,
}

impl Context {
    pub(crate) fn add_header(&mut self, tokens: proc_macro2::TokenStream) -> anyhow::Result<()> {
        self.header_items.push(syn::parse2(tokens)?);
        Ok(())
    }

    pub(crate) fn add_item(&mut self, tokens: proc_macro2::TokenStream) -> anyhow::Result<()> {
        let item = syn::parse2(tokens)?;
        self.body_items.push(item);
        Ok(())
    }

    pub(crate) fn to_string(&self) -> String {
        let items = vec![self.header_items.clone(), self.body_items.clone()].concat();
        let file = syn::File{
            shebang: None,
            attrs: vec![],
            items,
        };
        prettyplease::unparse(&file)
    }
}

pub(crate) type TokenResult = anyhow::Result<proc_macro2::TokenStream>;
