pub struct Store {

}

impl Store {
    fn get(&self, key: &[u8]) -> zeropb::Result<*const u8> {
        todo!()
    }

    fn set(&self, key: &[u8], value: &[u8]) -> zeropb::Result<()> {
        todo!()
    }

    fn delete(&self, key: &[u8]) -> zeropb::Result<()> {
        todo!()
    }

    fn has(&self, key: &[u8]) -> zeropb::Result<bool> {
        todo!()
    }
}