use crate::{Module, Server};

pub struct MockApp  {

}

impl MockApp {
    pub fn new() -> Self {
        MockApp {

        }
    }

    pub fn add_module<T: Module>(&mut self, name: &str, module: T, config: T::Config) {
    }

    pub fn add_mock_server<T: Server>(&mut self, server: T) {

    }
}