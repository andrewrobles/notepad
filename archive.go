package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// archiveMarkdown appends markdown to <home>/<directory>/<relPath>, creating
// any missing parent directories, then commits the change to the directory's
// git repository (initializing one if needed). It returns the full path of
// the archive file written to.
func archiveMarkdown(directory, relPath, markdown string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}

	dir := filepath.Join(home, directory)
	path := filepath.Join(dir, relPath)

	if err := writeArchive(path, markdown); err != nil {
		return "", err
	}

	if err := verifyArchive(path, markdown); err != nil {
		return "", err
	}

	if err := commitArchive(dir); err != nil {
		return "", err
	}

	return path, nil
}

func writeArchive(path, markdown string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create archive dir: %w", err)
	}

	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(markdown); err != nil {
		return fmt.Errorf("write archive: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync archive: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close archive: %w", err)
	}

	return nil
}

func verifyArchive(filename, markdown string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read archive for verification: %w", err)
	}

	if !strings.HasSuffix(string(data), markdown) {
		return fmt.Errorf("archive verification failed")
	}

	return nil
}

func commitArchive(dir string) error {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("inspect archive repository: %w", err)
		}
		if err := runGit(dir, "init"); err != nil {
			return err
		}
	}

	if err := runGit(dir, "add", "-A"); err != nil {
		return err
	}

	return runGit(dir, "commit", "--allow-empty-message", "-m", "")
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, output)
	}

	return nil
}
