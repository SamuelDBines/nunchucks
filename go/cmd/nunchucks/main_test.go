package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
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
