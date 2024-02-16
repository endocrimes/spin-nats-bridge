use spin_sdk::http::{IntoResponse, Request, Response};
use spin_sdk::http_component;

/// A simple Spin HTTP component.
#[http_component]
fn handle_message_logger(req: Request) -> anyhow::Result<impl IntoResponse> {
    println!("Processing Message: {:?}", String::from_utf8(req.body().to_vec()).unwrap_or("".into()));
    Ok(Response::builder()
        .status(200)
        .header("content-type", "text/plain")
        .build())
}
