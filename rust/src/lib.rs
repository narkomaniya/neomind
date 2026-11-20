//! neomind-core — ядро NeoMind на Rust.
//!
//! Никаких внешних зависимостей: хранилище заметок, JSON-answered команды
//! и аккуратная работа с путями/ошибками. Go-CLI только рисует UX.

use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::PathBuf;
use std::time::{SystemTime, UNIX_EPOCH};

pub const VERSION: &str = "0.1.0";

/// Папка данных: %USERPROFILE%/.neomind на Windows, ~/.neomind на POSIX.
pub fn data_dir() -> PathBuf {
    let home = std::env::var_os("USERPROFILE")
        .or_else(|| std::env::var_os("HOME"))
        .map(PathBuf::from)
        .unwrap_or_else(|| PathBuf::from("."));
    home.join(".neomind")
}

pub fn notes_path() -> PathBuf {
    data_dir().join("notes.txt")
}

/// Minimal JSON string escaping (quotes, backslashes, control chars).
pub fn json_escape(s: &str) -> String {
    let mut out = String::with_capacity(s.len() + 2);
    for c in s.chars() {
        match c {
            '"' => out.push_str("\\\""),
            '\\' => out.push_str("\\\\"),
            '\n' => out.push_str("\\n"),
            '\r' => out.push_str("\\r"),
            '\t' => out.push_str("\\t"),
            c if (c as u32) < 0x20 => out.push_str(&format!("\\u{:04x}", c as u32)),
            c => out.push(c),
        }
    }
    out
}

#[derive(Debug, Clone)]
pub struct Note {
    pub ts: u64,
    pub text: String,
}

fn now_secs() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_secs())
        .unwrap_or(0)
}

/// Читает все заметки. Битые строки молча пропускаются — файл мог править
/// человек или старая версия приложения.
pub fn note_list() -> Vec<Note> {
    let raw = match fs::read_to_string(notes_path()) {
        Ok(v) => v,
        Err(_) => return Vec::new(),
    };
    raw.lines()
        .filter_map(|line| {
            let (ts, text) = line.split_once('\t')?;
            Some(Note {
                ts: ts.parse().ok()?,
                text: text.to_string(),
            })
        })
        .collect()
}

/// Добавляет заметку, возвращает новое количество записей.
pub fn note_add(text: &str) -> Result<usize, String> {
    let clean: String = text
        .trim()
        .chars()
        .map(|c| if c == '\n' || c == '\r' { ' ' } else { c })
        .collect();
    if clean.is_empty() {
        return Err("note is empty".into());
    }
    if clean.chars().count() > 2000 {
        return Err("note is too long (max 2000 characters)".into());
    }
    fs::create_dir_all(data_dir()).map_err(|e| format!("cannot create data dir: {e}"))?;
    let mut file = OpenOptions::new()
        .create(true)
        .append(true)
        .open(notes_path())
        .map_err(|e| format!("cannot open notes: {e}"))?;
    writeln!(file, "{}\t{}", now_secs(), clean)
        .map_err(|e| format!("cannot write note: {e}"))?;
    Ok(note_list().len())
}

pub fn note_clear() -> Result<(), String> {
    match fs::remove_file(notes_path()) {
        Ok(()) => Ok(()),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(()),
        Err(e) => Err(format!("cannot clear notes: {e}")),
    }
}

pub fn notes_json() -> String {
    let notes = note_list();
    let items: Vec<String> = notes
        .iter()
        .map(|n| {
            format!(
                "{{\"ts\":{},\"text\":\"{}\"}}",
                n.ts,
                json_escape(&n.text)
            )
        })
        .collect();
    format!(
        "{{\"ok\":true,\"operation\":\"note_list\",\"notes\":[{}],\"count\":{}}}",
        items.join(","),
        notes.len()
    )
}

pub fn status_json() -> String {
    format!(
        "{{\"ok\":true,\"version\":\"{}\",\"core\":\"rust\",\"cli\":\"go\",\"count\":{},\"data_dir\":\"{}\"}}",
        VERSION,
        note_list().len(),
        json_escape(&data_dir().to_string_lossy())
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn escapes_json() {
        assert_eq!(json_escape("a\"b\\c\nd"), "a\\\"b\\\\c\\nd");
    }

    #[test]
    fn status_is_jsonish() {
        let s = status_json();
        assert!(s.contains("\"core\":\"rust\""));
        assert!(s.contains("\"version\":\"0.1.0\""));
    }
}
