# Contributing

## Code reviews
All submissions, including submissions by project members, require review. We
use Github pull requests for this purpose.

## Extending the schema

When adding new enumerations in the FlatBuffers schema:
- Append new values just before the "count" sentinel (e.g., `*_COUNT` entries).
- Update the C "wire" translation tables under c/src/wire/ (compile-time asserts help detect drift).
- Regenerate FlatBuffers code for both C (flatcc) and Go (flatc), then rebuild.

---

## Development

- Format C code:
  - `make -C c fmt`
  - `make -C c fmt-check`
- Re-generate C readers:
  - Happens automatically when building the C library (`c/src/generated` from `fbs-schema/*.fbs`)
- Re-generate Go schema:
  - `make -C go generate-schema-headers`
  - or run: `make -C go examples`
