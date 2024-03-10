use std::path::Path;
use prost_build::{Service, ServiceGenerator};
use prost_types::FileDescriptorSet;

pub struct Config {
    pub prost_config: prost_build::Config,
}

impl Default for Config {
    fn default() -> Self {
        let mut prost_cfg = prost_build::Config::default();
        prost_cfg
            .service_generator(Box::new(Gen::default()))
            .file_descriptor_set_path("file_descriptor_set.bin")
            .include("_includes.rs");
        Self { prost_config }
    }
}

impl Config {
    pub fn compile_fds(&mut self, protos: FileDescriptorSet) -> std::io::Result<()> {
        self.prost_config.compile_fds(protos)
    }

    pub fn compile_protos(&mut self, protos: &[impl AsRef<Path>], includes: &[impl AsRef<Path>]) -> std::io::Result<()> {
        self.prost_config.compile_protos(protos, includes)
    }
}

#[derive(Default)]
struct Gen {}

impl ServiceGenerator for Gen {
    fn generate(&mut self, service: Service, buf: &mut String) {
        let mut svc_gen = ServiceGen::default();
        svc_gen.generate(service);
        let file = syn::File {
            shebang: None,
            attrs: vec![],
            items: svc_gen.items,
        };
        let out = prettyplease::unparse(&file);
        buf.push_str(&out)
    }
}

#[derive(Default)]
struct ServiceGen {
    items: Vec<syn::Item>,
}

impl ServiceGen {
    fn add(&mut self, item: syn::Item) {
        self.items.push(item);
    }

    fn generate(&mut self, service: Service) {
        todo!()
    }
}