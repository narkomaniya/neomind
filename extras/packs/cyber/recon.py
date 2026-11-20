#!/usr/bin/env python3
"""Small, safe recon helper for labs you own.

Usage: python recon.py example.com
"""
from __future__ import annotations

import argparse
import socket
import sys


def resolve(host: str) -> list[str]:
    try:
        infos = socket.getaddrinfo(host, None)
    except socket.gaierror as exc:
        print(f"cannot resolve {host}: {exc}")
        return []
    return sorted({info[4][0] for info in infos})


def main() -> int:
    parser = argparse.ArgumentParser(description="Resolve a host and print its addresses.")
    parser.add_argument("host", help="hostname to inspect")
    args = parser.parse_args()

    addresses = resolve(args.host)
    if not addresses:
        return 1
    print(f"{args.host}:")
    for address in addresses:
        print(f"  {address}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
