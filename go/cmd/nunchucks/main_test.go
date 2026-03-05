package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	os.Stderr = w
	defer func() {
		os.Stderr = original
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	return buf.String()
}

func TestPrintHelpRender(t *testing.T) {
	output := captureStderr(t, func() {
		if err := printHelp("render"); err != nil {
			t.Fatalf("printHelp(render): %v", err)
		}
	})

	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected usage header, got %q", output)
	}
	if !strings.Contains(output, "render [options]") {
		t.Fatalf("expected render command usage, got %q", output)
	}
	if !strings.Contains(output, "-template string") {
		t.Fatalf("expected render flags, got %q", output)
	}
}

func TestPrintHelpUnknownTopic(t *testing.T) {
	err := printHelp("missing")
	if err == nil {
		t.Fatal("expected error for unknown help topic")
	}
	if !strings.Contains(err.Error(), "unknown help topic") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseDataEmpty(t *testing.T) {
	data, err := parseData("")
	if err != nil {
		t.Fatalf("parseData empty: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected empty map, got %#v", data)
	}
}

func TestParseDataInvalidJSON(t *testing.T) {
	_, err := parseData("{")
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestDiffTemplateStateDetectsAddsChangesAndDeletes(t *testing.T) {
	now := time.Unix(100, 0)
	before := map[string]fileState{
		"changed.njk": {modTime: now, size: 10},
		"removed.njk": {modTime: now, size: 20},
	}
	after := map[string]fileState{
		"added.njk":   {modTime: now, size: 30},
		"changed.njk": {modTime: now.Add(time.Second), size: 10},
	}

	got := diffTemplateState(before, after)
	want := []string{"added.njk", "changed.njk", "removed.njk"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("diffTemplateState() = %v, want %v", got, want)
	}
}

func TestRemoveDeletedOutputsRemovesMissingTemplates(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "pages", "index.njk")
	if err := os.MkdirAll(filepath.Dir(outFile), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(outFile, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	previous := map[string]fileState{
		"pages/index.njk": {modTime: time.Unix(100, 0), size: 5},
	}
	current := map[string]fileState{}

	if err := removeDeletedOutputs(dir, previous, current); err != nil {
		t.Fatalf("removeDeletedOutputs(): %v", err)
	}
	if _, err := os.Stat(outFile); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be removed, stat err = %v", outFile, err)
	}
}
