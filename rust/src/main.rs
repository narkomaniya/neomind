//! Тонкий протокол для Go-CLI и GUI: одна команда -> одна JSON-строка.

use neomind_core::{
    json_escape, note_add, note_clear, note_delete, note_list, note_update, notes_json,
    status_json, VERSION,
};

fn ok(op: &str, extra: &str) -> String {
    if extra.is_empty() {
        format!("{{\"ok\":true,\"operation\":\"{}\"}}", op)
    } else {
        format!("{{\"ok\":true,\"operation\":\"{}\",{}}}", op, extra)
    }
}

fn fail(op: &str, err: &str) -> String {
    format!(
        "{{\"ok\":false,\"operation\":\"{}\",\"error\":\"{}\"}}",
        op,
        json_escape(err)
    )
}

fn parse_id(s: Option<&String>) -> Result<u64, String> {
    s.and_then(|v| v.parse().ok())
        .ok_or_else(|| "invalid note id".to_string())
}

fn main() {
    let args: Vec<String> = std::env::args().skip(1).collect();
    let op = args.first().map(String::as_str).unwrap_or("status");
    let line = match op {
        "version" => ok("version", &format!("\"version\":\"{}\"", VERSION)),
        "status" => status_json(),
        "note_add" => match note_add(&args[1..].join(" ")) {
            Ok(id) => ok(
                "note_add",
                &format!("\"id\":{},\"count\":{}", id, note_list().len()),
            ),
            Err(e) => fail("note_add", &e),
        },
        "note_list" => notes_json(),
        "note_update" => match parse_id(args.get(1)) {
            Ok(id) => match note_update(id, &args[2..].join(" ")) {
                Ok(()) => ok("note_update", &format!("\"id\":{}", id)),
                Err(e) => fail("note_update", &e),
            },
            Err(e) => fail("note_update", &e),
        },
        "note_delete" => match parse_id(args.get(1)).and_then(|id| note_delete(id).map(|_| id)) {
            Ok(id) => ok("note_delete", &format!("\"id\":{}", id)),
            Err(e) => fail("note_delete", &e),
        },
        "note_clear" => match note_clear() {
            Ok(()) => ok("note_clear", ""),
            Err(e) => fail("note_clear", &e),
        },
        other => fail(other, "unknown core command"),
    };
    println!("{}", line);
}
