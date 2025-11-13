# Mort Monorepo

This repository hosts three tightly related projects that revolve around converting, browsing, and distributing mortality table data in XTbml form.

| Component | Language | Purpose |
|-----------|----------|---------|
| `internal/xtbml`, `cmd/xtbmlconvert` | Go | Conversion tooling that turns XTbml XML into JSON and other downstream-friendly formats. |
| `web/` | TypeScript | Web UI for discovering, searching, and exporting mortality tables. |
| `tui/` | Go | Terminal interface mirroring the web UX for offline-heavy workflows. |

## Development Principles

- **Test-first**: Every change starts with a failing test, followed immediately by the smallest implementation needed for green builds.
- **No excess work**: Only implement behavior that is directly exercised by tests.
- **Best practices**: Favor idiomatic Go (e.g., small packages, clear error handling) and modern TypeScript (ESM, strict compiler options).

## Getting Started

```sh
# Go tooling
go test ./...

# Web tooling (from ./web)
npm install
npm test
```

The `xml/` directory contains the source mortality tables and remains the single source of truth for conversions.
