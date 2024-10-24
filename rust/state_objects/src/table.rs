pub trait TableRow {

}

pub struct Table<Row> {
    _phantom: std::marker::PhantomData<Row>,
}