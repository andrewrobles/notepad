package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildNotepadBinary compiles the CLI once per test run and returns its path.
func buildNotepadBinary(t *testing.T) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "notepad")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = mustGetwd(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build notepad binary: %v\n%s", err, out)
	}

	return bin
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return wd
}

func runCLI(t *testing.T, bin, home string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "HOME="+home)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("run %s %v: %v", bin, args, err)
		}
	}

	return outBuf.String(), errBuf.String(), exitCode
}

func TestCLI_NoArgs_PrintsUsageAndExits1(t *testing.T) {
	bin := buildNotepadBinary(t)
	home := t.TempDir()

	stdout, _, code := runCLI(t, bin, home)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stdout, "usage: notepad save") {
		t.Errorf("stdout = %q, want it to contain usage text", stdout)
	}
}

func TestCLI_UnknownArg_PrintsUsageAndExits1(t *testing.T) {
	bin := buildNotepadBinary(t)
	home := t.TempDir()

	stdout, _, code := runCLI(t, bin, home, "bogus")

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stdout, "usage: notepad save") {
		t.Errorf("stdout = %q, want it to contain usage text", stdout)
	}
}

func TestCLI_Version(t *testing.T) {
	bin := buildNotepadBinary(t)
	home := t.TempDir()

	for _, flag := range []string{"-v", "--version"} {
		stdout, _, code := runCLI(t, bin, home, flag)
		if code != 0 {
			t.Errorf("%s: exit code = %d, want 0", flag, code)
		}
		if strings.TrimSpace(stdout) != version {
			t.Errorf("%s: stdout = %q, want %q", flag, strings.TrimSpace(stdout), version)
		}
	}
}

func TestCLI_Archive_EndToEnd(t *testing.T) {
	bin := buildNotepadBinary(t)
	home := t.TempDir()
	writeNotepad(t, home, "notes from the CLI test")

	stdout, _, code := runCLI(t, bin, home, "save")
	if code != 0 {
		t.Fatalf("archive exit code = %d, stdout=%q", code, stdout)
	}
	if !strings.Contains(stdout, "Archived notepad to") {
		t.Errorf("stdout = %q, want it to mention archiving", stdout)
	}

	mdPath := filepath.Join(home, "notepad", ".md")
	data, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read .md: %v", err)
	}
	if len(data) != 0 {
		t.Errorf(".md not cleared after archive, got %q", data)
	}

	// Running archive again with nothing to archive should succeed with the "no contents" message.
	stdout, _, code = runCLI(t, bin, home, "save")
	if code != 0 {
		t.Fatalf("second archive exit code = %d, stdout=%q", code, stdout)
	}
	if !strings.Contains(stdout, "No notepad contents to archive.") {
		t.Errorf("stdout = %q, want no-contents message", stdout)
	}
}

func TestCLI_Archive_WithTitle(t *testing.T) {
	bin := buildNotepadBinary(t)
	home := t.TempDir()
	writeNotepad(t, home, "titled CLI notes")

	stdout, _, code := runCLI(t, bin, home, "save", "ideas")
	if code != 0 {
		t.Fatalf("archive exit code = %d, stdout=%q", code, stdout)
	}
	if !strings.Contains(stdout, "Archived notepad to") {
		t.Errorf("stdout = %q, want it to mention archiving", stdout)
	}

	archived, err := os.ReadFile(archivePath(home, "ideas"))
	if err != nil {
		t.Fatalf("read titled archive file: %v", err)
	}
	if string(archived) != "titled CLI notes" {
		t.Errorf("archive content = %q, want %q", archived, "titled CLI notes")
	}
}

func TestCLI_Archive_TooManyArgs_PrintsUsageAndExits1(t *testing.T) {
	bin := buildNotepadBinary(t)
	home := t.TempDir()

	stdout, _, code := runCLI(t, bin, home, "save", "title", "extra")

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stdout, "usage: notepad save") {
		t.Errorf("stdout = %q, want it to contain usage text", stdout)
	}
}
