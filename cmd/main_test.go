package main

import (
	"os"
	"path"
	"testing"
)

func TestSaveData(t *testing.T) {
	dir := t.TempDir()
	p := path.Join(dir, "kvs")
	data := []byte("hello world")

	t.Logf("writing data to: %s", p)
	err := SaveData(p, data)
	if err != nil {
		t.Fatal(err) // Fatal stops the test immediately — no point continuing if the write itself failed
	}

	// actual verification: does the written content match what we expect?
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(data) {
		t.Errorf("content mismatch: expected %q, got %q", data, got)
	} else {
		t.Logf("✓ content matches exactly: %q", got)
	}
}

func TestSaveData2(t *testing.T) {
	dir := t.TempDir()
	p := path.Join(dir, "kvs")
	data := []byte("hello world")

	err := SaveData2(p, data)
	if err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(data) {
		t.Errorf("content mismatch: expected %q, got %q", data, got)
	}

	// extra check: did the tmp file actually disappear after Rename?
	entries, _ := os.ReadDir(dir)
	t.Logf("remaining files in dir: %d (should be 1: kvs)", len(entries))
	for _, e := range entries {
		t.Logf("  - %s", e.Name())
	}
}

func TestSaveData3(t *testing.T) {
	dir := t.TempDir()
	p := path.Join(dir, "kvs")
	data := []byte("hello world")

	err := SaveData3(p, data)
	if err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(data) {
		t.Errorf("content mismatch: expected %q, got %q", data, got)
	} else {
		t.Logf("✓ SaveData3 succeeded with fsync, content: %q", got)
	}
}

func TestLogAppend(t *testing.T) {
	dir := t.TempDir()
	p := path.Join(dir, "kvs")

	fp, err := LogCreate(p)
	if err != nil {
		t.Fatal(err)
	}

	// write multiple lines to confirm Append actually appends, not overwrites
	lines := []string{"line1", "line2", "line3"}
	for _, line := range lines {
		if err := LogAppend(fp, line); err != nil {
			t.Fatal(err)
		}
	}
	fp.Close()

	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}

	expected := "line1\nline2\nline3\n"
	t.Logf("actual file content:\n%s", string(got))

	if string(got) != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, string(got))
	} else {
		t.Log("✓ Append works correctly: each line was added at the end without overwriting the previous ones")
	}
}
