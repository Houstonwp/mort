package jsonupdate

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"mort/internal/xtbml"
)

func TestUpdateTracksChangedSources(t *testing.T) {
	root := newRepository(t)
	writeFixture(t, root, "xml/deleted.xml", "Deleted")
	writeFixture(t, root, "xml/modified.xml", "Original")
	writeFixture(t, root, "xml/renamed.xml", "Renamed")
	for _, name := range []string{"deleted", "modified", "renamed"} {
		generate(t, root, name)
	}
	commit(t, root)
	remove(t, filepath.Join(root, "xml/deleted.xml"))
	writeFixture(t, root, "xml/modified.xml", "Modified")
	runGit(t, root, "mv", "xml/renamed.xml", "xml/moved.xml")
	writeFixture(t, root, "xml/nested/new table\nname.xml", "New")
	writeFile(t, root, "xml/ignored.txt", []byte("not XML"))
	writeFile(t, root, "json/changelog_state.json", []byte(`{"last_log_ms":123}`))
	var output bytes.Buffer
	if err := Update(root, "HEAD", false, &output); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"modified", "moved", "nested/new table\nname"} {
		assertGenerated(t, root, name)
	}
	for _, name := range []string{"deleted", "renamed"} {
		if _, err := os.Stat(filepath.Join(root, "json", name+".json")); !os.IsNotExist(err) {
			t.Errorf("obsolete %s.json still exists: %v", name, err)
		}
	}
	if got := string(readFile(t, root, "json/changelog_state.json")); got != `{"last_log_ms":123}` {
		t.Errorf("state changed: %s", got)
	}
	if err := Update(root, "HEAD", true, &output); err != nil {
		t.Fatalf("check after regeneration: %v", err)
	}
}

func TestCheckRejectsStaleOrMissingOutputWithoutWriting(t *testing.T) {
	for _, outputExists := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "stale"}[outputExists], func(t *testing.T) {
			root := newRepository(t)
			writeFixture(t, root, "xml/new.xml", "New")
			if outputExists {
				writeFile(t, root, "json/new.json", []byte("stale JSON"))
			}
			before := runGit(t, root, "status", "--porcelain", "--untracked-files=all")
			var output bytes.Buffer
			if err := Update(root, "HEAD", true, &output); err == nil {
				t.Fatal("check should fail for stale or missing JSON")
			}
			if got := runGit(t, root, "status", "--porcelain", "--untracked-files=all"); got != before {
				t.Fatalf("check modified worktree: before %q, after %q", before, got)
			}
			if outputExists && string(readFile(t, root, "json/new.json")) != "stale JSON" {
				t.Fatal("check rewrote stale output")
			}
		})
	}
}

func TestCheckDeletedSource(t *testing.T) {
	root := newRepository(t)
	writeFixture(t, root, "xml/deleted.xml", "Deleted")
	generate(t, root, "deleted")
	commit(t, root)
	remove(t, filepath.Join(root, "xml/deleted.xml"))
	var output bytes.Buffer
	if err := Update(root, "HEAD", true, &output); err == nil {
		t.Fatal("check should reject leftover JSON for deleted XML")
	}
	if err := Update(root, "HEAD", false, &output); err != nil {
		t.Fatal(err)
	}
	if err := Update(root, "HEAD", true, &output); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateValidatesCommittedAndJSONOnlyChanges(t *testing.T) {
	root := newRepository(t)
	base := strings.TrimSpace(runGit(t, root, "rev-parse", "HEAD"))
	writeFixture(t, root, "xml/table.xml", "Table")
	generate(t, root, "table")
	commit(t, root)
	var output bytes.Buffer
	if err := Update(root, base, true, &output); err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, "json/table.json", []byte("tampered"))
	if err := Update(root, "HEAD", true, &output); err == nil {
		t.Fatal("check should reject a JSON-only edit")
	}
	if err := Update(root, "HEAD", false, &output); err != nil {
		t.Fatal(err)
	}
	assertGenerated(t, root, "table")
}

func TestUpdateNoChangesAndInvalidInputs(t *testing.T) {
	root := newRepository(t)
	var output bytes.Buffer
	if err := Update(root, "HEAD", true, &output); err != nil {
		t.Fatal(err)
	}
	if err := Update(root, "missing-ref", false, &output); err == nil {
		t.Fatal("expected invalid base ref to fail")
	}
	writeFile(t, root, "xml/broken.xml", []byte("<XTbML>"))
	if err := Update(root, "HEAD", false, &output); err == nil {
		t.Fatal("expected malformed XML to fail")
	}
	if _, err := os.Stat(filepath.Join(root, "json/broken.json")); !os.IsNotExist(err) {
		t.Fatal("invalid XML created output")
	}
}

func newRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init", "-q")
	runGit(t, root, "config", "user.name", "Test")
	runGit(t, root, "config", "user.email", "test@example.invalid")
	runGit(t, root, "commit", "-q", "--allow-empty", "-m", "Initial fixture")
	return root
}

func runGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	output, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
	return string(output)
}

func commit(t *testing.T, root string) {
	t.Helper()
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-q", "-m", "Fixture")
}

func writeFile(t *testing.T, root, name string, data []byte) {
	t.Helper()
	file := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, root, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeFixture(t *testing.T, root, name, tableName string) {
	t.Helper()
	data, err := os.ReadFile("../xtbml/testdata/table_small.xml")
	if err != nil {
		t.Fatal(err)
	}
	data = bytes.ReplaceAll(data, []byte("Sample Table"), []byte(tableName))
	writeFile(t, root, name, data)
}

func generate(t *testing.T, root, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := xtbml.ConvertFile(filepath.Join(root, "xml", name+".xml"), filepath.Join(root, "json", name+".json")); err != nil {
		t.Fatal(err)
	}
}

func assertGenerated(t *testing.T, root, name string) {
	t.Helper()
	want, err := xtbml.ConvertXTbml(bytes.NewReader(readFile(t, root, "xml/"+name+".xml")))
	if err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, root, "json/"+name+".json"); !bytes.Equal(got, want) {
		t.Errorf("%s.json does not match its XML source", name)
	}
}

func remove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateRejectsNonRegularFiles(t *testing.T) {
	for _, name := range []string{"xml/table.xml", "json/table.json"} {
		t.Run(name, func(t *testing.T) {
			root := newRepository(t)
			writeFixture(t, root, "xml/table.xml", "Table")
			if name == "xml/table.xml" {
				remove(t, filepath.Join(root, name))
			} else if err := os.MkdirAll(filepath.Join(root, "json"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("../outside.xml", filepath.Join(root, name)); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			if err := Update(root, "HEAD", false, &output); err == nil {
				t.Fatal("should reject symbolic links")
			}
		})
	}
}

func TestInvalidXMLPreventsAllWrites(t *testing.T) {
	root := newRepository(t)
	writeFixture(t, root, "xml/a-valid.xml", "Valid")
	writeFile(t, root, "xml/z-invalid.xml", []byte("<XTbML>"))
	var output bytes.Buffer
	if err := Update(root, "HEAD", false, &output); err == nil {
		t.Fatal("expected malformed XML to fail")
	}
	if _, err := os.Stat(filepath.Join(root, "json/a-valid.json")); !os.IsNotExist(err) {
		t.Fatal("failed batch partially wrote generated JSON")
	}
}
