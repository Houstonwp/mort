# Mort Monorepo

Mort combines the tooling we use to ingest XTbML mortality tables, the web app that lets researchers browse them, and a terminal UI for power users. Everything shares the same data and tests so we can evolve the ecosystem together.

## Contents

- [Requirements](#requirements)
- [Getting Started](#getting-started)
- [Converter CLI](#converter-cli)
- [Web App](#web-app)
- [Terminal UI](#terminal-ui)
- [Data Source](#data-source)
- [Verification Checklist](#verification-checklist)

## Requirements

- Go 1.25+
- Node.js 20+ (with npm)
- git

## Getting Started

```sh
git clone https://github.com/Houstonwp/mort.git
cd mort
go test ./...          # sanity check Go toolchain
cd web && npm install  # install UI deps (run once)
```

## Converter CLI

- Source lives in `internal/xtbml/` with the executable in `cmd/xtbmlconvert/`.
- Written in Go 1.25 with table-driven tests and fixtures scoped to the package.
- Convert an XML file to JSON:

  ```sh
  go run ./cmd/xtbmlconvert -in xml/sample.xml -out json/sample.json
  ```

- Run converter-specific tests (from repo root):

  ```sh
  go test ./cmd/... ./internal/xtbml/...
  ```

## Web App

- Located in `web/` and built with TypeScript, Preact, and Vite.
- Install deps once: `npm install`
- Start the dev server with hot reload:

  ```sh
  npm run dev
  # Navigate to the printed localhost URL.
  ```

- Execute unit tests: `npm test`
- The app consumes the converted JSON files emitted by the Go tooling; drop fixtures under `web/src/testdata` when needed.

## Terminal UI

- Located in `tui/` and mirrors common web flows for keyboard-heavy or offline usage.
- Run interactively from the repo root:

  ```sh
  go run ./tui
  ```

- Tests live beside the packages under `tui/`. Run them with:

  ```sh
  go test ./tui/...
  ```

## Data Source

- Canonical mortality tables live under `xml/`.
- Never commit raw mortality data outside this directory; generated JSON fixtures belong in `json/` or package-local `testdata/` folders.

## Verification Checklist

Run these commands before opening a PR:

```sh
go test ./...      # validates converter + tui + shared libraries
cd web && npm test # validates the web UI
```
