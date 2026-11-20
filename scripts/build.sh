#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist="$root/dist"
mkdir -p "$dist"

echo '==> building Rust core'
cargo build --release --manifest-path "$root/rust/Cargo.toml"

echo '==> building Go CLI'
cd "$root"
go build -trimpath -o "$dist/neomind" ./cmd/neomind
cp -f "$root/rust/target/release/neomind-core" "$dist/neomind-core"
chmod +x "$dist/neomind" "$dist/neomind-core"

echo "NeoMind is ready: $dist/neomind"
echo 'Add that folder to PATH or run: ./dist/neomind help'
