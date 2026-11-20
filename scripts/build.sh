#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist="$root/dist"
mkdir -p "$dist"

echo '==> building Rust core'
cargo build --release --manifest-path "$root/rust/Cargo.toml"

echo '==> building Go CLI'
cd "$root"
CGO_ENABLED=0 go build -trimpath -o "$dist/neomind" ./cmd/neomind
cp -f "$root/rust/target/release/neomind-core" "$dist/neomind-core"

# The native GUI needs GTK + WebKit headers. Build it only when they exist;
# CLI and core always build cleanly on a bare machine.
if [ "${NEOMIND_SKIP_GUI:-0}" != "1" ] && pkg-config --exists gtk+-3.0 webkit2gtk-4.0 2>/dev/null; then
    echo '==> building desktop GUI'
    CGO_ENABLED=1 go build -trimpath -o "$dist/neomind-gui" ./cmd/neomind-gui
else
    echo '==> skipping GUI (install libgtk-3-dev + libwebkit2gtk-4.0-dev to enable)'
fi

chmod +x "$dist/neomind" "$dist/neomind-core"
[ -f "$dist/neomind-gui" ] && chmod +x "$dist/neomind-gui" || true

echo "NeoMind is ready: $dist/neomind"
echo 'Add that folder to PATH or run: ./dist/neomind help'
