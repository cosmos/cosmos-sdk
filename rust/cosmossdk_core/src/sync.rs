extern crate core;
use core::ops::Fn;
use crate::{Context, Result};

pub struct PrepareContext;

pub struct ExecContext(Context);

pub type Exec<T> = Result<dyn Fn(&mut ExecContext) -> Result<T>>;

pub type Completer<R> = dyn Fn(&mut ExecContext) -> Result<R>;

pub type Completer1<P1, R> = dyn Fn(&mut ExecContext, P1) -> Result<R>;
