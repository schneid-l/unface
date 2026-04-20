package main

import (
	"go/format"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGenerateBasic(t *testing.T) {
	req := Request{
		Dir:   "testdata/basic",
		Types: []string{"URL", "Counter", "Event", "Flags"},
		Mode:  "dispatch",
	}
	got, err := Generate(req)
	if err != nil {
		t.Fatal(err)
	}

	// The generated output must be gofmt-clean.
	if _, err := format.Source(got); err != nil {
		t.Fatalf("generated code does not parse: %v\nsource:\n%s", err, got)
	}

	s := string(got)
	wants := []string{
		"func (recv *URL) Unface(src any) error",
		"case string:",
		"return recv.Unstring(v)",
		"if m, ok := unface.MapOf(src); ok",
		"return recv.Unmap(m)",

		"func (recv *Counter) Unface(src any) error",
		"if n, ok := unface.NumberOf(src); ok",
		"return recv.Unnumber(n)",

		"func (recv *Event) Unface(src any) error",
		"case time.Time:",
		"return recv.Untime(v)",

		"func (recv *Flags) Unface(src any) error",
		"case bool:",
		"return recv.Unbool(v)",
	}
	for _, w := range wants {
		if !strings.Contains(s, w) {
			t.Errorf("generated output missing %q\n--- output ---\n%s", w, s)
		}
	}
}

// Smoke-compile the generated code against the testdata package to prove
// it type-checks end-to-end.
func TestGenerateCompiles(t *testing.T) {
	req := Request{
		Dir:   "testdata/basic",
		Types: []string{"URL", "Counter", "Event", "Flags"},
		Mode:  "dispatch",
	}
	got, err := Generate(req)
	if err != nil {
		t.Fatal(err)
	}
	path := "testdata/basic/generated_test_unface.go"
	t.Cleanup(func() { _ = exec.Command("rm", "-f", path).Run() })

	formatted, err := format.Source(got)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeFile(path, formatted); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "vet", "./testdata/basic/...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go vet failed: %v\n%s", err, out)
	}
}

// TestGenerateBasicGolden pins the entire formatted output of dispatch
// mode against cmd/unfacegen/testdata/basic/basic.golden. Set
// UNFACEGEN_UPDATE=1 to refresh the golden after an intentional change.
func TestGenerateBasicGolden(t *testing.T) {
	req := Request{
		Dir:   "testdata/basic",
		Types: []string{"URL", "Counter", "Event", "Flags"},
		Mode:  "dispatch",
	}
	assertGolden(t, req, "testdata/basic/basic.golden")
}

// TestGenerateWalkerGolden pins the entire formatted output of walker
// mode against cmd/unfacegen/testdata/walker/walker.golden. Set
// UNFACEGEN_UPDATE=1 to refresh the golden after an intentional change.
func TestGenerateWalkerGolden(t *testing.T) {
	req := Request{
		Dir:   "testdata/walker",
		Types: []string{"Config"},
		Mode:  "walker",
	}
	assertGolden(t, req, "testdata/walker/walker.golden")
}

// TestGenerateWalkerCompiles verifies the walker-mode output type-checks
// when placed into its package.
func TestGenerateWalkerCompiles(t *testing.T) {
	req := Request{
		Dir:   "testdata/walker",
		Types: []string{"Config"},
		Mode:  "walker",
	}
	got, err := Generate(req)
	if err != nil {
		t.Fatal(err)
	}
	formatted, err := format.Source(got)
	if err != nil {
		t.Fatalf("generated walker code does not parse: %v\nsource:\n%s", err, got)
	}
	path := "testdata/walker/generated_test_unface.go"
	t.Cleanup(func() { _ = exec.Command("rm", "-f", path).Run() })
	if err := writeFile(path, formatted); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "vet", "./testdata/walker/...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go vet failed: %v\n%s", err, out)
	}
}

// assertGolden runs Generate and diffs the formatted output against the
// checked-in golden file. Use UNFACEGEN_UPDATE=1 to overwrite the golden
// when an intentional change lands.
func assertGolden(t *testing.T, req Request, goldenPath string) {
	t.Helper()
	got, err := Generate(req)
	if err != nil {
		t.Fatal(err)
	}
	formatted, err := format.Source(got)
	if err != nil {
		t.Fatalf("generated code does not parse: %v\nsource:\n%s", err, got)
	}
	if os.Getenv("UNFACEGEN_UPDATE") == "1" {
		if err := writeFile(goldenPath, formatted); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v", goldenPath, err)
	}
	if string(formatted) != string(want) {
		t.Fatalf("generated output diverged from %s\n"+
			"run UNFACEGEN_UPDATE=1 go test ./cmd/unfacegen/... to refresh.\n"+
			"--- got ---\n%s\n--- want ---\n%s",
			goldenPath, formatted, want)
	}
}

func writeFile(path string, data []byte) error {
	return execWrite(path, data)
}
