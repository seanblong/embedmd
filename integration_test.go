package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIntegration(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "sample/docs.md")
	got, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("could not process file (%v): %s", err, got)
	}
	wants, err := os.ReadFile(filepath.Join("sample", "result.md"))
	if err != nil {
		t.Fatalf("could not read result: %v", err)
	}
	if string(got) != string(wants) {
		t.Fatalf("got bad result (compared to result.md):\n%s", got)
	}
}
