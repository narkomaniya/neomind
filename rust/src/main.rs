//! Тонкий протокол для Go-CLI: одна команда -> одна JSON-строка.
//! Go не знает про Rust-типы, Rust не знает про UX — только JSON.

use neomind_core::{note_add, note_clear, notes_json, status_json, VERSION};

fn ok(op: &str, extra: &str) -> String {
    if extra.is_empty() {
        format!("{{\"ok\":true,\"operation\":\"{}\"}}", op)
    } else {
        format!("{{\"ok\":true,\"operation\":\"{}\",{} }}", op, extra)
    }
}

fn fail(op: &str, err: &str) -> String {
    format!(
        "{{\"ok\":false,\"operation\":\"{}\",\"error\":\"{}\"}}",
        op,
        neomind_core::json_escape(err)
    )
}

fn main() {
    let args: Vec<String> = std::env::args().skip(1).collect();
    let op = args.first().map(String::as_str).unwrap_or("status");
    let line = match op {
        "version" => ok("version", &format!("\"version\":\"{}\"", VERSION)),
        "status" => status_json(),
        "note_add" => {
            let text = args[1..].join(" ").trim().to_string();
            match note_add(&text) {
                Ok(n) => ok("note_add", &format!("\"count\":{}", n)),
                Err(e) => fail("note_add", &e),
            }
        }
        "note_list" => notes_json(),
        "note_clear" => match note_clear() {
            Ok(()) => ok("note_clear", ""),
            Err(e) => fail("note_clear", &e),
        },
        other => fail(other, "unknown core command"),
    };
    println!("{}", line);
}
