#!/usr/bin/env python3
# Remove BOM from collection config files
import os

files = [
    "/root/hlf-go/repo/collections/collection-config-electricity-de.json",
    "/root/hlf-go/repo/collections/collection-config-hydrogen-de.json"
]

for fpath in files:
    with open(fpath, "rb") as f:
        content = f.read()
    # Remove UTF-8 BOM if present
    if content.startswith(b'\xef\xbb\xbf'):
        content = content[3:]
        with open(fpath, "wb") as f:
            f.write(content)
        print(f"Removed BOM from {fpath}")
    else:
        print(f"No BOM in {fpath}")
