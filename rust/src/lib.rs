//! neomind-core — the Rust core of NeoMind.
//!
//! No external dependencies: note storage, JSON-answered commands and careful
//! path/error handling. The GUI and CLI only provide the UX.

use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::PathBuf;
use std::time::{SystemTime, UNIX_EPOCH};

pub const VERSION: &str = "0.5.3";

/// Data folder: %USERPROFILE%/.neomind on Windows, ~/.neomind on POSIX.
/// NEOMIND_DATA_DIR redirects storage (tests, portable mode).
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

/// Minimal JSON escaping for strings (quotes, slashes, control characters).
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

/// One note. id is epoch milliseconds; older records in seconds are detected
/// and converted to milliseconds when displayed.
#[derive(Debug, Clone)]
pub struct Note {
    pub id: u64,
    pub text: String,
    /// File name shown in the Folders tree. Empty for legacy notes.
    pub name: String,
    /// Internal folder path or an absolute path on disk. Empty for legacy notes.
    pub folder: String,
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

/// Store multiline text on a single file line.
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

/// Read all notes, sorted by id (oldest first).
/// Broken lines are skipped silently — a human may have edited the file.
pub fn note_list() -> Vec<Note> {
    let raw = match fs::read_to_string(notes_path()) {
        Ok(v) => v,
        Err(_) => return Vec::new(),
    };
    let mut notes: Vec<Note> = raw
        .lines()
        .filter_map(|line| {
            let mut parts = line.splitn(4, '\t');
            let id: u64 = parts.next()?.parse().ok()?;
            let first = parts.next()?;
            // New format: id<TAB>name<TAB>folder<TAB>text.
            // Old format: id<TAB>text — still readable without migration.
            let (name, folder, text) = match (parts.next(), parts.next()) {
                (Some(folder), Some(text)) => {
                    (decode_text(first), decode_text(folder), decode_text(text))
                }
                _ => (String::new(), String::new(), decode_text(first)),
            };
            Some(Note {
                id,
                text,
                name,
                folder,
            })
        })
        .collect();
    notes.sort_by_key(|n| n.id);
    notes
}

fn check_len(text: &str) -> Result<(), String> {
    if text.chars().count() > 20_000 {
        return Err("file is too long (max 20000 characters)".into());
    }
    Ok(())
}

fn clean_text(text: &str) -> Result<String, String> {
    let clean = text.trim().to_string();
    if clean.is_empty() {
        return Err("note is empty".into());
    }
    check_len(&clean)?;
    Ok(clean)
}

/// Code files keep their exact formatting — no trimming.
fn file_text(text: &str) -> Result<String, String> {
    check_len(text)?;
    Ok(text.to_string())
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
            writeln!(
                file,
                "{}\t{}\t{}\t{}",
                n.id,
                encode_text(&n.name),
                encode_text(&n.folder),
                encode_text(&n.text)
            )
            .map_err(|e| format!("cannot write note: {e}"))?;
        }
        file.flush().ok();
    }
    fs::rename(&tmp, notes_path()).map_err(|e| format!("cannot replace notes: {e}"))
}

fn next_id(notes: &[Note]) -> u64 {
    let mut id = now_millis();
    if let Some(last) = notes.last() {
        if id <= last.id {
            id = last.id.saturating_add(1);
        }
    }
    id
}

fn append_note(note: &Note) -> Result<(), String> {
    fs::create_dir_all(data_dir()).map_err(|e| format!("cannot create data dir: {e}"))?;
    let mut file = OpenOptions::new()
        .create(true)
        .append(true)
        .open(notes_path())
        .map_err(|e| format!("cannot open notes: {e}"))?;
    writeln!(
        file,
        "{}\t{}\t{}\t{}",
        note.id,
        encode_text(&note.name),
        encode_text(&note.folder),
        encode_text(&note.text)
    )
    .map_err(|e| format!("cannot write note: {e}"))
}

/// Add a legacy note and return its id.
pub fn note_add(text: &str) -> Result<u64, String> {
    let clean = clean_text(text)?;
    let notes = note_list();
    let id = next_id(&notes);
    append_note(&Note {
        id,
        text: clean,
        name: String::new(),
        folder: String::new(),
    })?;
    Ok(id)
}

/// Add a named file, optionally inside an internal or on-disk folder.
pub fn note_add_file(name: &str, folder: &str, text: &str) -> Result<u64, String> {
    let name = name.trim();
    if name.is_empty() {
        return Err("file name is empty".into());
    }
    let text = file_text(text)?;
    let notes = note_list();
    let id = next_id(&notes);
    append_note(&Note {
        id,
        text,
        name: name.to_string(),
        folder: folder.to_string(),
    })?;
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

/// Update a file's content, name and folder without reformatting its text.
pub fn note_update_file(id: u64, name: &str, folder: &str, text: &str) -> Result<(), String> {
    let name = name.trim();
    if name.is_empty() {
        return Err("file name is empty".into());
    }
    let text = file_text(text)?;
    let mut notes = note_list();
    let found = notes.iter_mut().find(|n| n.id == id);
    match found {
        Some(n) => {
            n.name = name.to_string();
            n.folder = folder.to_string();
            n.text = text;
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
        "{{\"id\":{},\"created\":{},\"name\":\"{}\",\"folder\":\"{}\",\"text\":\"{}\"}}",
        n.id,
        n.created_millis(),
        json_escape(&n.name),
        json_escape(&n.folder),
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
        let raw = "first line\nsecond\tline\\end";
        assert_eq!(decode_text(&encode_text(raw)), raw);
    }

    #[test]
    fn status_is_jsonish() {
        let s = status_json();
        assert!(s.contains("\"core\":\"rust\""));
        assert!(s.contains("\"version\":\"0.5.0\""));
    }
}
