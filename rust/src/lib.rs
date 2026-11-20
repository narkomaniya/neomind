//! neomind-core — ядро NeoMind на Rust.
//!
//! Никаких внешних зависимостей: хранилище заметок, JSON-answered команды
//! и аккуратная работа с путями/ошибками. GUI и CLI только рисуют UX.

use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::PathBuf;
use std::time::{SystemTime, UNIX_EPOCH};

pub const VERSION: &str = "0.3.0";

/// Папка данных: %USERPROFILE%/.neomind на Windows, ~/.neomind на POSIX.
/// NEOMIND_DATA_DIR позволяет перенаправить хранилище (тесты, портативный режим).
pub fn data_dir() -> PathBuf {
    if let Some(p) = std::env::var_os("NEOMIND_DATA_DIR") {
        return PathBuf::from(p);
    }
    let home = std::env::var_os("USERPROFILE")
        .or_else(|| std::env::var_os("HOME"))
        .map(PathBuf::from)
        .unwrap_or_else(|| PathBuf::from("."));
    home.join(".neomind")
}

pub fn notes_path() -> PathBuf {
    data_dir().join("notes.txt")
}

/// Минимальный JSON-escape для строк (кавычки, слеши, управляющие символы).
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

/// Одна заметка. id — миллисекунды epoch; старые записи в секундах
/// распознаются и превращаются в миллисекунды при показе.
#[derive(Debug, Clone)]
pub struct Note {
    pub id: u64,
    pub text: String,
}

impl Note {
    pub fn created_millis(&self) -> u64 {
        if self.id < 1_000_000_000_000 {
            self.id.saturating_mul(1000)
        } else {
            self.id
        }
    }
}

/// Сохраняем многострочный текст в одной строке файла.
fn encode_text(s: &str) -> String {
    let mut out = String::with_capacity(s.len());
    for c in s.chars() {
        match c {
            '\\' => out.push_str("\\\\"),
            '\n' => out.push_str("\\n"),
            '\r' => out.push_str("\\r"),
            '\t' => out.push_str("\\t"),
            c => out.push(c),
        }
    }
    out
}

fn decode_text(s: &str) -> String {
    let mut out = String::with_capacity(s.len());
    let mut it = s.chars();
    while let Some(c) = it.next() {
        if c == '\\' {
            match it.next() {
                Some('n') => out.push('\n'),
                Some('r') => out.push('\r'),
                Some('t') => out.push('\t'),
                Some('\\') => out.push('\\'),
                Some(other) => {
                    out.push('\\');
                    out.push(other);
                }
                None => out.push('\\'),
            }
        } else {
            out.push(c);
        }
    }
    out
}

fn now_millis() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_millis() as u64)
        .unwrap_or(0)
}

/// Читает все заметки, сортирует по id (старые сверху).
/// Битые строки молча пропускаются — файл мог править человек.
pub fn note_list() -> Vec<Note> {
    let raw = match fs::read_to_string(notes_path()) {
        Ok(v) => v,
        Err(_) => return Vec::new(),
    };
    let mut notes: Vec<Note> = raw
        .lines()
        .filter_map(|line| {
            let (id, text) = line.split_once('\t')?;
            Some(Note {
                id: id.parse().ok()?,
                text: decode_text(text),
            })
        })
        .collect();
    notes.sort_by_key(|n| n.id);
    notes
}

fn clean_text(text: &str) -> Result<String, String> {
    let clean = text.trim().to_string();
    if clean.is_empty() {
        return Err("note is empty".into());
    }
    if clean.chars().count() > 20_000 {
        return Err("note is too long (max 20000 characters)".into());
    }
    Ok(clean)
}

fn write_all(notes: &[Note]) -> Result<(), String> {
    fs::create_dir_all(data_dir()).map_err(|e| format!("cannot create data dir: {e}"))?;
    let tmp = data_dir().join("notes.tmp");
    {
        let mut file = OpenOptions::new()
            .create(true)
            .write(true)
            .truncate(true)
            .open(&tmp)
            .map_err(|e| format!("cannot open temp notes: {e}"))?;
        for n in notes {
            writeln!(file, "{}\t{}", n.id, encode_text(&n.text))
                .map_err(|e| format!("cannot write note: {e}"))?;
        }
        file.flush().ok();
    }
    fs::rename(&tmp, notes_path()).map_err(|e| format!("cannot replace notes: {e}"))
}

/// Добавляет заметку, возвращает её id.
pub fn note_add(text: &str) -> Result<u64, String> {
    let clean = clean_text(text)?;
    let mut notes = note_list();
    let mut id = now_millis();
    if let Some(last) = notes.last() {
        if id <= last.id {
            id = last.id.saturating_add(1);
        }
    }
    fs::create_dir_all(data_dir()).map_err(|e| format!("cannot create data dir: {e}"))?;
    let mut file = OpenOptions::new()
        .create(true)
        .append(true)
        .open(notes_path())
        .map_err(|e| format!("cannot open notes: {e}"))?;
    writeln!(file, "{}\t{}", id, encode_text(&clean))
        .map_err(|e| format!("cannot write note: {e}"))?;
    notes.push(Note { id, text: clean });
    Ok(id)
}

pub fn note_update(id: u64, text: &str) -> Result<(), String> {
    let clean = clean_text(text)?;
    let mut notes = note_list();
    let found = notes.iter_mut().find(|n| n.id == id);
    match found {
        Some(n) => {
            n.text = clean;
            write_all(&notes)
        }
        None => Err(format!("note {id} not found")),
    }
}

pub fn note_delete(id: u64) -> Result<(), String> {
    let mut notes = note_list();
    let before = notes.len();
    notes.retain(|n| n.id != id);
    if notes.len() == before {
        return Err(format!("note {id} not found"));
    }
    write_all(&notes)
}

pub fn note_clear() -> Result<(), String> {
    match fs::remove_file(notes_path()) {
        Ok(()) => Ok(()),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(()),
        Err(e) => Err(format!("cannot clear notes: {e}")),
    }
}

fn note_json(n: &Note) -> String {
    format!(
        "{{\"id\":{},\"created\":{},\"text\":\"{}\"}}",
        n.id,
        n.created_millis(),
        json_escape(&n.text)
    )
}

pub fn notes_json() -> String {
    let notes = note_list();
    let items: Vec<String> = notes.iter().map(note_json).collect();
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
    fn multiline_text_roundtrip() {
        let raw = "первая строка\nвторая\tстрока\\конец";
        assert_eq!(decode_text(&encode_text(raw)), raw);
    }

    #[test]
    fn status_is_jsonish() {
        let s = status_json();
        assert!(s.contains("\"core\":\"rust\""));
        assert!(s.contains("\"version\":\"0.3.0\""));
    }
}
