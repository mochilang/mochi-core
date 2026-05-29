package golden

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

// Runner defines a test processor function.
type Runner func(srcPath string) ([]byte, error)

// Run executes golden file-based tests.
func Run(t *testing.T, dir, srcExt, goldenExt string, fn Runner) {
	t.Helper()
	rootDir := findRepoRoot(t)
	pattern := filepath.Join(rootDir, dir, "*"+srcExt)
	if only := os.Getenv("MOCHI_ROSETTA_ONLY"); only != "" {
		pattern = filepath.Join(rootDir, dir, only+srcExt)
	}

	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("failed to list %s files in %s: %v", srcExt, dir, err)
	}
	if len(files) == 0 {
		t.Fatalf("no test files found: %s", pattern)
	}
	sort.Strings(files)

	for _, src := range files {
		src := src
		name := strings.TrimSuffix(filepath.Base(src), srcExt)
		wantPath := filepath.Join(rootDir, dir, name+goldenExt)

		t.Run(name, func(t *testing.T) {
			got, err := fn(src)
			if err != nil {
				t.Fatalf("process error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}

			got = normalizeOutput(rootDir, got)

			if *update {
				if err := os.WriteFile(wantPath, got, 0644); err != nil {
					t.Fatalf("failed to write golden: %v", err)
				}
				t.Logf("updated: %s", wantPath)
				return
			}

			want, err := os.ReadFile(wantPath)
			if err != nil {
				t.Fatalf("failed to read golden: %v", err)
			}
			want = normalizeOutput(rootDir, want)

			if !bytes.Equal(got, want) {
				t.Errorf("golden mismatch for %s\n\n--- Got ---\n%s\n\n--- Want ---\n%s\n", name+goldenExt, got, want)
			}
		})
	}
}

// RunWithSummary works like Run but logs a summary at the end.
func RunWithSummary(t *testing.T, dir, srcExt, goldenExt string, fn Runner) {
	t.Helper()
	rootDir := findRepoRoot(t)
	pattern := filepath.Join(rootDir, dir, "*"+srcExt)

	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("failed to list %s files in %s: %v", srcExt, dir, err)
	}
	if len(files) == 0 {
		t.Fatalf("no test files found: %s", pattern)
	}
	sort.Strings(files)

	var passed, failed int
	for _, src := range files {
		src := src
		name := strings.TrimSuffix(filepath.Base(src), srcExt)
		wantPath := filepath.Join(rootDir, dir, name+goldenExt)

		ok := t.Run(name, func(t *testing.T) {
			got, err := fn(src)
			if err != nil {
				t.Fatalf("process error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			got = normalizeOutput(rootDir, got)
			if *update {
				if err := os.WriteFile(wantPath, got, 0644); err != nil {
					t.Fatalf("failed to write golden: %v", err)
				}
				return
			}
			want, err := os.ReadFile(wantPath)
			if err != nil {
				t.Fatalf("failed to read golden: %v", err)
			}
			want = normalizeOutput(rootDir, want)
			if !bytes.Equal(got, want) {
				t.Errorf("golden mismatch for %s\n\n--- Got ---\n%s\n\n--- Want ---\n%s\n", name+goldenExt, got, want)
			}
		})
		if ok {
			passed++
		} else {
			failed++
		}
	}
	t.Logf("Summary: %d passed, %d failed", passed, failed)
}

// RunFirstFailure executes golden tests and stops after the first failure.
func RunFirstFailure(t *testing.T, dir, srcExt, goldenExt string, fn Runner) {
	t.Helper()
	rootDir := findRepoRoot(t)
	pattern := filepath.Join(rootDir, dir, "*"+srcExt)
	if only := os.Getenv("MOCHI_ROSETTA_ONLY"); only != "" {
		pattern = filepath.Join(rootDir, dir, only+srcExt)
	}

	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("failed to list %s files in %s: %v", srcExt, dir, err)
	}
	if len(files) == 0 {
		t.Fatalf("no test files found: %s", pattern)
	}
	sort.Strings(files)

	for _, src := range files {
		src := src
		name := strings.TrimSuffix(filepath.Base(src), srcExt)
		wantPath := filepath.Join(rootDir, dir, name+goldenExt)

		ok := t.Run(name, func(t *testing.T) {
			got, err := fn(src)
			if err != nil {
				t.Fatalf("process error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			got = normalizeOutput(rootDir, got)
			if *update {
				if err := os.WriteFile(wantPath, got, 0644); err != nil {
					t.Fatalf("failed to write golden: %v", err)
				}
				return
			}
			want, err := os.ReadFile(wantPath)
			if err != nil {
				t.Fatalf("failed to read golden: %v", err)
			}
			want = normalizeOutput(rootDir, want)
			if !bytes.Equal(got, want) {
				t.Errorf("golden mismatch for %s\n\n--- Got ---\n%s\n\n--- Want ---\n%s\n", name+goldenExt, got, want)
			}
		})
		if !ok {
			t.Fatalf("first failing program: %s", name)
		}
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal("cannot determine working directory")
	}
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("go.mod not found (not in Go module)")
	return ""
}

func normalizeOutput(root string, b []byte) []byte {
	out := string(b)
	out = strings.ReplaceAll(out, "\r\n", "\n")
	out = strings.ReplaceAll(out, `\`, "/")
	out = strings.ReplaceAll(out, filepath.ToSlash(root)+"/", "")
	out = strings.ReplaceAll(out, filepath.ToSlash(root), "")
	out = strings.ReplaceAll(out, "github.com/mochi-lang/mochi/", "")
	out = strings.ReplaceAll(out, "mochi/tests/", "tests/")
	durRE := regexp.MustCompile(`\([0-9]+(\.[0-9]+)?(ns|µs|ms|s)\)`)
	out = durRE.ReplaceAllString(out, "(X)")
	tsRE := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`)
	out = tsRE.ReplaceAllString(out, "2006-01-02T15:04:05Z")
	out = strings.TrimSpace(out)
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return []byte(out)
}
