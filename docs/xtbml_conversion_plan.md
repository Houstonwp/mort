# XTbml XML → JSON Conversion TDD Plan

## Goals
- Transform every mortality table XML in `xml/` into normalized JSON residing in `json/`.
- Guarantee deterministic identifiers, metadata, and data payloads suitable for both web and TUI clients.
- Keep the surfacing CLI (`cmd/xtbmlconvert`) tiny by delegating logic to `internal/xtbml`.

## Data & Test Assets
- Treat `xml/` as read-only fixtures. Copy representative samples into `internal/xtbml/testdata/` for automated tests.
- Add golden JSON snapshots to `internal/xtbml/testdata/json/`. Each test commits both the source XML and its expected JSON.

## Phase 1 – Schema Discovery Helpers
1. **Failing test**: `TestInferVersion` ensures we can read the `XTbml` version attribute. Create XML fixture missing the attribute to assert error paths.
2. **Implementation**: Add `InferVersion(r io.Reader) (string, error)` using `encoding/xml` streaming decoder.

## Phase 2 – Table Metadata Parsing
1. **Failing tests**:
   - `TestParseTableMetadata_Minimal` uses a tiny XML with name, publisher, and effective date.
   - `TestParseTableMetadata_InvalidDate` validates error formatting.
2. **Implementation**: Introduce `ParseTableMetadata` returning a Go struct, reusing `NormalizeIdentifier`.

## Phase 3 – Rate Grid Extraction
1. **Failing test**: `TestParseRates` verifies that cells map age/duration pairs correctly; include missing-age scenario to confirm validation.
2. **Implementation**: Stream XML rows into a dense representation (e.g., `map[string]RateRow` or slice).

## Phase 4 – JSON Serialization Contract
1. **Failing golden test**: `TestMarshalTable_Golden` loads XML, converts to JSON, and compares bytes with `testdata/json/<table>.json`.
2. **Implementation**: Compose `ConvertXTbml(r io.Reader) ([]byte, error)` that stitches metadata + rates into the JSON schema consumed by UI layers.
3. Use `go-cmp` or custom diff for readable failures; record newline-terminated JSON with stable ordering.

## Phase 5 – Filesystem Workflow
1. **Failing integration test**: `TestConvertDirectory` writes XML fixtures into a temp dir, runs `ConvertDirectory(src, dst)`, and asserts:
   - JSON files land in `dst/<table>.json`.
   - Non-XML files are ignored.
2. **Implementation**: `internal/xtbml/filesystem.go` walks directories, concurrent-safe but deterministic (sorted input list).

## Phase 6 – CLI Wrapper
1. **Failing test**: Use `cmd/xtbmlconvert/main_test.go` with `exec.Command` to run `go test` subcommand or, if main package cannot be tested directly, wrap logic in `internal/cmd/runner` and use tests there. Check:
   - Exit code 0 on success.
   - Exit code 1 and stderr message when conversion fails.
2. **Implementation**: CLI accepts `--src=xml --dst=json` defaults, ensures `json/` exists, and logs per-file completion.

## Continuous Verification
- Add `go test ./...` to CI; include a `scripts/convert.sh` that fails when JSON outputs are stale (compare timestamps or `git diff`).
- Whenever new XML arrives, first add a failing golden test before updating the converter.
