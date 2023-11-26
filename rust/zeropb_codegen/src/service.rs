use crate::ctx::Context;
use crate::r#type::gen_message_name;
use heck::ToSnakeCase;
use prost_types::{MethodDescriptorProto, ServiceDescriptorProto};
use std::fmt::Write;

pub(crate) fn gen_service(
    service: &ServiceDescriptorProto,
    ctx: &mut Context,
) -> anyhow::Result<()> {
    gen_service_server(service, ctx)?;
    // gen_service_client(service, ctx)?;
    Ok(())
}

fn gen_service_server(service: &ServiceDescriptorProto, ctx: &mut Context) -> anyhow::Result<()> {
    write!(
        ctx,
        "trait {}Server<Ctx> {{\n",
        service.name.clone().unwrap()
    )?;
    for method in service.method.iter() {
        gen_server_method(method, ctx)?;
    }
    write!(ctx, "}}\n\n")?;
    Ok(())
}

pub(crate) fn gen_server_method(
    method: &MethodDescriptorProto,
    ctx: &mut Context,
) -> anyhow::Result<()> {
    let method_name = method.name.clone().unwrap().to_snake_case();
    let input_type = method.input_type.clone().unwrap();
    let output_type = method.output_type.clone().unwrap();
    write!(ctx, "    fn {}(&self, ctx: &Ctx, req: &", method_name)?;
    gen_message_name(&input_type, ctx)?;
    write!(ctx, ") -> Result<");
    gen_message_name(&output_type, ctx)?;
    write!(ctx, ", zeropb::Status>;\n")?;
    Ok(())
}

fn gen_service_client(service: &ServiceDescriptorProto, ctx: &mut Context) -> anyhow::Result<()> {
    write!(
        ctx,
        "struct {}Client<'a, Client> {{\n",
        service.name.clone().unwrap()
    )?;
    write!(ctx, "   client_conn: Client,\n")?;
    for method in service.method.iter() {
        let method_name = method.name.clone().unwrap().to_snake_case();
        write!(ctx, "   {}: zeropb::Handler<'a, i32, i32>,\n", method_name)?;
    }
    write!(ctx, "}}\n\n")?;
    write!(
        ctx,
        "impl <'a, Client: ClientConn> {}Client<'a, Client> {{\n",
        service.name.clone().unwrap()
    )?;
    write!(ctx, "    pub fn new(client_conn: Client) -> Self {{")?;
    write!(ctx, "        Self {{")?;
    write!(ctx, "            client_conn,\n")?;
    for method in service.method.iter() {
        let method_name = method.name.clone().unwrap();
        let method_name_snake = method_name.to_snake_case();
        write!(
            ctx,
            "            {}: client_conn.resolve_unary(\"{}\"),\n",
            method_name_snake, method_name
        )?;
    }
    write!(ctx, "        }}\n")?;
    write!(ctx, "    }}\n\n")?;
    for method in service.method.iter() {
        gen_client_method(method, ctx)?;
    }
    write!(ctx, "}}\n\n")?;
    Ok(())
}

pub(crate) fn gen_client_method(
    method: &MethodDescriptorProto,
    ctx: &mut Context,
) -> anyhow::Result<()> {
    let method_name = method.name.clone().unwrap().to_snake_case();
    let input_type = method.input_type.clone().unwrap();
    let output_type = method.output_type.clone().unwrap();
    write!(ctx, "    fn {}(&self, req: &", method_name)?;
    gen_message_name(&input_type, ctx)?;
    write!(ctx, ") -> Result<")?;
    gen_message_name(&output_type, ctx)?;
    write!(ctx, ", zeropb::Status> {{\n")?;
    write!(ctx, "        unimplemented!()\n")?;
    write!(ctx, "    }}\n")?;
    Ok(())
}
