#!/usr/bin/env python3
"""Summarize the sample dataset with only the Python standard library."""
from __future__ import annotations

import csv
import statistics
from pathlib import Path


def load(path: Path) -> list[dict[str, str]]:
    with path.open(newline="", encoding="utf-8") as handle:
        return list(csv.DictReader(handle))


def main() -> None:
    rows = load(Path(__file__).with_name("dataset.csv"))
    scores = [float(row["score"]) for row in rows]
    minutes = [float(row["minutes"]) for row in rows]
    print(f"rows: {len(rows)}")
    print(f"average score: {statistics.mean(scores):.2f}")
    print(f"median minutes: {statistics.median(minutes):.1f}")
    best = max(rows, key=lambda row: float(row["score"]))
    print(f"best run: {best['name']} ({best['score']})")


if __name__ == "__main__":
    main()
