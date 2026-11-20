# Snippet shelf

Keep short, reusable pieces close to your notes.

## JavaScript

```js
export const debounce = (fn, ms = 160) => {
  let timer;
  return (...args) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), ms);
  };
};
```

## Python

```python
from pathlib import Path

def read_lines(path: str) -> list[str]:
    return Path(path).read_text(encoding="utf-8").splitlines()
```

## Rust

```rust
pub fn clamp(value: i64, low: i64, high: i64) -> i64 {
    value.max(low).min(high)
}
```

## Shell

```powershell
Get-ChildItem -Recurse -Filter *.rs | Select-String "TODO"
```
