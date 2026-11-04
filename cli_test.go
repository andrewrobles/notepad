package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// buildNotepadBinary compiles the CLI once per test run and returns its path.
func buildNotepadBinary(t *testing.T) string {
	t.Helper()

	bin := exeName(filepath.Join(t.TempDir(), "notepad"))
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

// exeName appends the platform executable suffix so built binaries can be run.
func exeName(path string) string {
	if runtime.GOOS == "windows" {
		return path + ".exe"
	}
	return path
}

// buildFakeNotepad compiles a stand-in executable that records its invocation
// by creating the file named in NOTEPAD_MARKER, so passthrough can be tested
// without launching the real GUI editor.
func buildFakeNotepad(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	src := filepath.Join(dir, "fake.go")
	const prog = `package main

import "os"

func main() {
	if m := os.Getenv("NOTEPAD_MARKER"); m != "" {
		os.WriteFile(m, []byte("called"), 0644)
	}
}
`
	if err := os.WriteFile(src, []byte(prog), 0644); err != nil {
		t.Fatalf("write fake notepad source: %v", err)
	}

	bin := exeName(filepath.Join(dir, "fakenotepad"))
	cmd := exec.Command("go", "build", "-o", bin, src)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fake notepad: %v\n%s", err, out)
	}

	return bin
}

// waitForFile waits for path to exist, accommodating the detached (Start-only)
// launch of the forwarded notepad process.
func waitForFile(t *testing.T, path string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("file %s was not created within timeout", path)
}

func runCLI(t *testing.T, bin, home string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)

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

func TestCLI_NoArgs_ForwardsToNotepad(t *testing.T) {
	bin := buildNotepadBinary(t)
	home := t.TempDir()
	fake := buildFakeNotepad(t)
	marker := filepath.Join(t.TempDir(), "marker")
	t.Setenv("NOTEPAD_EXE", fake)
	t.Setenv("NOTEPAD_MARKER", marker)

	_, _, code := runCLI(t, bin, home)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	waitForFile(t, marker)
}

func TestCLI_UnknownArg_ForwardsToNotepad(t *testing.T) {
	bin := buildNotepadBinary(t)
	home := t.TempDir()
	fake := buildFakeNotepad(t)
	marker := filepath.Join(t.TempDir(), "marker")
	t.Setenv("NOTEPAD_EXE", fake)
	t.Setenv("NOTEPAD_MARKER", marker)

	_, _, code := runCLI(t, bin, home, "bogus")

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	waitForFile(t, marker)
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

	stdout, _, code := runCLI(t, bin, home, "archive")
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
	stdout, _, code = runCLI(t, bin, home, "archive")
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

	stdout, _, code := runCLI(t, bin, home, "archive", "ideas")
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

	stdout, _, code := runCLI(t, bin, home, "archive", "title", "extra")

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stdout, "usage: notepad archive") {
		t.Errorf("stdout = %q, want it to contain usage text", stdout)
	}
}
