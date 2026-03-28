package abi

import (
	"testing"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

func TestLoadFromFile_ReturnsNotImplemented(t *testing.T) {
	_, err := LoadFromFile("test.json")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "not implemented" {
		t.Fatalf("expected 'not implemented', got %q", err.Error())
	}
}

func TestFindMethod_ReturnsNotImplemented(t *testing.T) {
	_, err := FindMethod(ethabi.ABI{}, "transfer")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "not implemented" {
		t.Fatalf("expected 'not implemented', got %q", err.Error())
	}
}

func TestListMethods_ReturnsNil(t *testing.T) {
	result := ListMethods(ethabi.ABI{})
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}
