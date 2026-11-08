package main

import (
	"testing"
)

func TestRunSuccess(t *testing.T) {
	err := run([]string{})
	if err != nil {
		t.Fatalf("expected run() to return nil, got: %v", err)
	}
}

func TestVersionDefaults(t *testing.T) {
	if Version == "" {
		t.Fatal("expected Version to be non-empty")
	}
}

