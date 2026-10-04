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
- [Data-update Pipeline](#data-update-pipeline)

## Requirements

- Go 1.25+
- Node.js 20+ (with npm)
- git

## Getting Started

```sh
git clone https://github.com/Houstonwp/mort.git
cd mort
go test ./...          # sanity check Go toolchain
cd web && npm ci       # install locked UI dependencies (run once)
```

## Converter CLI

- Source lives in `internal/xtbml/` with the executable in `cmd/xtbmlconvert/`.
- Written in Go 1.25 with table-driven tests and fixtures scoped to the package.
- Convert an XML directory to JSON:

  ```sh
  go run ./cmd/xtbmlconvert -src xml -dst json
  ```

- Run converter-specific tests (from repo root):

  ```sh
  go test ./cmd/... ./internal/xtbml/...
  ```

## Web App

- Located in `web/` and built with TypeScript, Preact, and Vite.
- Install locked dependencies: `npm ci`
- Start the dev server with hot reload:

  ```sh
  npm run dev
  # Navigate to the printed localhost URL.
  ```

- Validate the static site and generated detail routes: `npm run build`. There is no web unit-test script yet; a successful build is not a browser-interaction test.
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
go test ./...                   # converter + tui + data-update regression tests
go vet ./...                    # Go static checks
(cd web && npm ci && npm run build) # validate the static web build
```

## Data-update pipeline

- `go run ./cmd/updatejson` regenerates JSON only for changed XML/JSON paths relative to `HEAD`, including staged and untracked files. It handles added, modified, renamed, and deleted sources, and preserves `json/changelog_state.json`.
- For changes already committed on a feature branch, use `go run ./cmd/updatejson -base <base-commit>`. Commit the resulting JSON alongside the XML. Use `-check` to report stale or missing output without modifying files.
- Pull requests validate the proposed merge against the PR base commit. This is read-only for both fork and same-repository PRs; automation never pushes to a contributor's branch.
- The scheduled SOA sync runs tests, downloads updates, regenerates changed JSON, and builds the web app before committing XML, JSON, and the changelog cursor together. A failed conversion/build prevents the commit, so the next run can retry from the previous committed cursor.
- Syncs are serialized and manual syncs are restricted to the default branch. A manual table ID must be a positive integer. The sync retains its existing direct-default-branch push policy and needs repository rules to allow that bot push; it does not bypass branch protection.
- A successful data commit calls the existing Pages deployment workflow with its exact SHA. This explicit handoff is required because a push made with `GITHUB_TOKEN` does not trigger another push workflow. Deployments also run Go tests/vet before building and publishing.
- Run `go run ./cmd/changelogsync` only when you intend to contact SOA and update local data. The regression tests use temporary local Git repositories and fixtures; CI validation does not fetch upstream tables.
