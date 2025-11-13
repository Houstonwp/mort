# Repository Guidelines

## Project Structure & Module Organization
The repository is a monorepo centered on mortality tables in `xml/`. Go source lives under `cmd/` for binaries and `internal/xtbml/` for shared logic and tests. Future UI layers will sit in `web/` (TypeScript React) and `tui/` (Go). Keep fixtures inside the package under test or the `testdata/` folder to minimize coupling.

## Build, Test, and Development Commands
- `go test ./...` – runs every Go unit test; execute before every commit.
- `gofmt -w <files>` – formats Go code; required on touched files.
- `npm install && npm test` (within `web/`) – installs dependencies and executes UI tests once the frontend exists.

## Coding Style & Naming Conventions
Use Go modules with Go 1.25; rely on `gofmt` for spacing and imports. Package names stay short and lowercase (`xtbml`, `tui`). Exported identifiers must include doc comments, while private helpers should remain concise (<= 40 lines). For TypeScript, prefer ESM, `strict` mode, and descriptive file names like `tableSearch.ts`.

## Testing Guidelines
All work follows TDD: write a failing test first, then the minimal implementation. Go tests live alongside code using `_test.go` suffixes and `Test<Subject>` names. TypeScript tests will use Jest with `*.test.ts`. Favor table-driven tests for parsers and converters, and store reusable fixtures in `testdata/`. Target 100% coverage for new logic; justify gaps explicitly in PRs.

## Commit & Pull Request Guidelines
Commits should be small, scoped to a single behavior, and use imperative subjects (e.g., `Add normalize identifier helper`). PRs must:
1. Reference related issues or TODOs.
2. Describe the failing test added first and how subsequent commits make it pass.
3. Include screenshots or CLI output when UI behavior changes.
4. List verification commands (`go test`, `npm test`).

## Security & Configuration Tips
Never commit raw mortality data outside `xml/`. Secrets belong in local env vars or `.env.local`, not version control. Validate third-party dependencies with `go env GOPROXY=direct` or npm advisories before adoption.
