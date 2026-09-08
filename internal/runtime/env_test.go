package runtime

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "\n" +
		"# a comment\n" +
		"A=1\n" +
		"B=\"quoted\"\n" +
		"  C = unquoted  \n" +
		"NO_EQUALS\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	got, err := ReadEnvFile(path)
	if err != nil {
		t.Fatalf("ReadEnvFile() error = %v", err)
	}
	want := []string{"A=1", "B=quoted", "C=unquoted"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadEnvFile() = %#v, want %#v", got, want)
	}
}

func TestReadEnvFileMissingFile(t *testing.T) {
	got, err := ReadEnvFile(filepath.Join(t.TempDir(), "does-not-exist.env"))
	if err != nil {
		t.Fatalf("ReadEnvFile() error = %v, want nil", err)
	}
	if got != nil {
		t.Fatalf("ReadEnvFile() = %#v, want nil", got)
	}
}

func TestReadEnvFileEmptyPath(t *testing.T) {
	got, err := ReadEnvFile("")
	if err != nil {
		t.Fatalf("ReadEnvFile() error = %v, want nil", err)
	}
	if got != nil {
		t.Fatalf("ReadEnvFile() = %#v, want nil", got)
	}
}

func TestChildEnviron(t *testing.T) {
	t.Setenv("ANGEE_TEST_CHILD_ENVIRON", "inherited")
	fileEnv := []string{"A=1", "B=2"}
	got := ChildEnviron(fileEnv)
	if len(got) < len(fileEnv) {
		t.Fatalf("ChildEnviron() = %#v, want at least %d entries", got, len(fileEnv))
	}
	// The file entries come last, in order, so os/exec's last-wins semantics
	// let them override any inherited copy.
	tail := got[len(got)-len(fileEnv):]
	if !reflect.DeepEqual(tail, fileEnv) {
		t.Fatalf("ChildEnviron() tail = %#v, want %#v", tail, fileEnv)
	}
	// The inherited environment is present ahead of the appended entries.
	var sawInherited bool
	for _, entry := range got[:len(got)-len(fileEnv)] {
		if entry == "ANGEE_TEST_CHILD_ENVIRON=inherited" {
			sawInherited = true
			break
		}
	}
	if !sawInherited {
		t.Fatalf("ChildEnviron() did not include the inherited environment: %#v", got)
	}
	// The result must not alias the input's backing array: mutating the input
	// after the call must not change the returned slice.
	if len(fileEnv) > 0 && len(got) > 0 && &got[len(got)-len(fileEnv)] == &fileEnv[0] {
		t.Fatal("ChildEnviron() aliased the input slice's backing array")
	}
	fileEnv[0] = "A=mutated"
	if tail[0] != "A=1" {
		t.Fatalf("ChildEnviron() result changed when input mutated: %q", tail[0])
	}
}
