package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// writeNotepad creates ~/notepad/.md under the given fake home with the given contents.
func writeNotepad(t *testing.T, home, contents string) {
	t.Helper()

	dir := filepath.Join(home, "notepad")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir notepad dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, ".md"), []byte(contents), 0644); err != nil {
		t.Fatalf("write .md: %v", err)
	}
}

// archivePath returns the expected archive path for today's date and the
// given title (empty for the untitled ".md" archive).
func archivePath(home, title string) string {
	date := time.Now().Format("2006-01-02")
	return filepath.Join(home, "notepad", date, title+".md")
}

func commitCount(t *testing.T, dir string) int {
	t.Helper()

	out, err := exec.Command("git", "-C", dir, "rev-list", "--count", "HEAD").Output()
	if err != nil {
		t.Fatalf("git rev-list: %v", err)
	}

	n := strings.TrimSpace(string(out))
	count, err := strconv.Atoi(n)
	if err != nil {
		t.Fatalf("parse commit count %q: %v", n, err)
	}

	return count
}

func TestArchiveNotepad_Success(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeNotepad(t, home, "hello world")

	if err := archiveNotepad(""); err != nil {
		t.Fatalf("archiveNotepad: %v", err)
	}

	notepadDir := filepath.Join(home, "notepad")

	// Source file should be cleared.
	data, err := os.ReadFile(filepath.Join(notepadDir, ".md"))
	if err != nil {
		t.Fatalf("read .md: %v", err)
	}
	if len(data) != 0 {
		t.Errorf(".md not cleared, got %q", data)
	}

	// Untitled archives land in <date>/.md.
	archived, err := os.ReadFile(archivePath(home, ""))
	if err != nil {
		t.Fatalf("read archive file: %v", err)
	}
	if string(archived) != "hello world" {
		t.Errorf("archive content = %q, want %q", archived, "hello world")
	}

	// A git repo should have been initialized and committed to.
	if _, err := os.Stat(filepath.Join(notepadDir, ".git")); err != nil {
		t.Fatalf("expected .git dir: %v", err)
	}
	if got := commitCount(t, notepadDir); got != 1 {
		t.Errorf("commit count = %d, want 1", got)
	}
}

func TestArchiveNotepad_WithTitle(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeNotepad(t, home, "titled content")

	if err := archiveNotepad("ideas"); err != nil {
		t.Fatalf("archiveNotepad: %v", err)
	}

	archived, err := os.ReadFile(archivePath(home, "ideas"))
	if err != nil {
		t.Fatalf("read titled archive file: %v", err)
	}
	if string(archived) != "titled content" {
		t.Errorf("archive content = %q, want %q", archived, "titled content")
	}

	// The untitled archive file should not have been created.
	if _, err := os.Stat(archivePath(home, "")); !os.IsNotExist(err) {
		t.Errorf("expected no untitled archive file, err=%v", err)
	}
}

func TestArchiveNotepad_Empty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeNotepad(t, home, "")

	if err := archiveNotepad(""); err != nil {
		t.Fatalf("archiveNotepad: %v", err)
	}

	// No archive should have been created, and no git repo initialized.
	if _, err := os.Stat(filepath.Join(home, "notepad", ".git")); !os.IsNotExist(err) {
		t.Errorf("expected no .git dir to be created for empty notepad, err=%v", err)
	}
}

func TestArchiveNotepad_MissingFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// Note: ~/notepad/.md deliberately not created.

	err := archiveNotepad("")
	if err == nil {
		t.Fatal("expected error for missing notepad file, got nil")
	}
	if !strings.Contains(err.Error(), "open notepad") {
		t.Errorf("error = %v, want it to mention 'open notepad'", err)
	}
}

func TestArchiveNotepad_SameDayArchivesAppend(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeNotepad(t, home, "first entry")

	if err := archiveNotepad(""); err != nil {
		t.Fatalf("first archiveNotepad: %v", err)
	}

	// .md is now empty; archiving again should be a no-op, not an error.
	if err := archiveNotepad(""); err != nil {
		t.Fatalf("second archiveNotepad: %v", err)
	}

	notepadDir := filepath.Join(home, "notepad")
	if got := commitCount(t, notepadDir); got != 1 {
		t.Errorf("commit count after no-op second run = %d, want 1", got)
	}

	// Writing new content and archiving again on the same day should append
	// to the existing <date>/.md archive file rather than fail or overwrite.
	writeNotepad(t, home, "second entry")
	if err := archiveNotepad(""); err != nil {
		t.Fatalf("third archiveNotepad: %v", err)
	}
	if got := commitCount(t, notepadDir); got != 2 {
		t.Errorf("commit count after second archive = %d, want 2", got)
	}

	archived, err := os.ReadFile(archivePath(home, ""))
	if err != nil {
		t.Fatalf("read archive file: %v", err)
	}
	want := "first entry" + "second entry"
	if string(archived) != want {
		t.Errorf("archive content = %q, want %q", archived, want)
	}
}

func TestArchiveNotepad_TitledAndUntitledAreSeparate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeNotepad(t, home, "untitled entry")
	if err := archiveNotepad(""); err != nil {
		t.Fatalf("archiveNotepad(untitled): %v", err)
	}

	writeNotepad(t, home, "titled entry")
	if err := archiveNotepad("ideas"); err != nil {
		t.Fatalf("archiveNotepad(ideas): %v", err)
	}

	untitled, err := os.ReadFile(archivePath(home, ""))
	if err != nil {
		t.Fatalf("read untitled archive: %v", err)
	}
	if string(untitled) != "untitled entry" {
		t.Errorf("untitled archive = %q, want %q", untitled, "untitled entry")
	}

	titled, err := os.ReadFile(archivePath(home, "ideas"))
	if err != nil {
		t.Fatalf("read titled archive: %v", err)
	}
	if string(titled) != "titled entry" {
		t.Errorf("titled archive = %q, want %q", titled, "titled entry")
	}
}
