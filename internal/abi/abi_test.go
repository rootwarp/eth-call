package abi

import (
	"strings"
	"testing"
)

func TestLoadFromFile_ValidABI(t *testing.T) {
	parsed, err := LoadFromFile("../../test/data/erc20.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify methods map is non-empty
	if len(parsed.Methods) == 0 {
		t.Fatal("expected non-empty methods map")
	}

	// Verify known ERC-20 methods exist
	expectedMethods := []string{"transfer", "approve", "balanceOf", "allowance", "totalSupply"}
	for _, name := range expectedMethods {
		if _, ok := parsed.Methods[name]; !ok {
			t.Errorf("expected method %q in parsed ABI", name)
		}
	}
}

func TestLoadFromFile_FileNotFound(t *testing.T) {
	_, err := LoadFromFile("nonexistent.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}

	// Error must contain the file path
	if !strings.Contains(err.Error(), "nonexistent.json") {
		t.Errorf("error should contain file path, got: %v", err)
	}

	// Error must use the abi: prefix
	if !strings.HasPrefix(err.Error(), "abi:") {
		t.Errorf("error should start with 'abi:', got: %v", err)
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	_, err := LoadFromFile("../../test/data/invalid.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}

	// Error must use the abi: prefix
	if !strings.HasPrefix(err.Error(), "abi:") {
		t.Errorf("error should start with 'abi:', got: %v", err)
	}

	// Error must mention invalid JSON ABI
	if !strings.Contains(err.Error(), "invalid JSON ABI") {
		t.Errorf("error should contain 'invalid JSON ABI', got: %v", err)
	}
}
