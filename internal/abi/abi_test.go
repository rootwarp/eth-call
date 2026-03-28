package abi

import (
	"strings"
	"testing"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
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

func loadERC20(t *testing.T) ethabi.ABI {
	t.Helper()
	parsed, err := LoadFromFile("../../test/data/erc20.json")
	if err != nil {
		t.Fatalf("failed to load ERC-20 ABI: %v", err)
	}
	return parsed
}

func TestFindMethod_ExistingMethod(t *testing.T) {
	parsed := loadERC20(t)

	method, err := FindMethod(parsed, "transfer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if method.RawName != "transfer" {
		t.Errorf("expected RawName 'transfer', got %q", method.RawName)
	}

	// transfer(address,uint256) has 2 inputs
	if len(method.Inputs) != 2 {
		t.Errorf("expected 2 inputs, got %d", len(method.Inputs))
	}
}

func TestFindMethod_NotFound(t *testing.T) {
	parsed := loadERC20(t)

	_, err := FindMethod(parsed, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing method, got nil")
	}

	// Error must use the abi: prefix
	if !strings.HasPrefix(err.Error(), "abi:") {
		t.Errorf("error should start with 'abi:', got: %v", err)
	}

	// Error must contain the method name
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should contain method name, got: %v", err)
	}

	// Error must list available methods
	if !strings.Contains(err.Error(), "transfer") {
		t.Errorf("error should list available methods, got: %v", err)
	}
}

func TestFindMethod_EmptyName(t *testing.T) {
	parsed := loadERC20(t)

	_, err := FindMethod(parsed, "")
	if err == nil {
		t.Fatal("expected error for empty method name, got nil")
	}
}

func TestFindMethod_OverloadedByRawName(t *testing.T) {
	// ABI with two overloaded "foo" methods
	abiJSON := `[
		{"type":"function","name":"foo","inputs":[{"name":"a","type":"uint256"}],"outputs":[]},
		{"type":"function","name":"foo","inputs":[{"name":"a","type":"uint256"},{"name":"b","type":"uint256"}],"outputs":[]}
	]`
	parsed, err := ethabi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	// "foo" should resolve via direct map key (first overload)
	method, err := FindMethod(parsed, "foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method.RawName != "foo" {
		t.Errorf("expected RawName 'foo', got %q", method.RawName)
	}

	// "foo0" should resolve via direct map key (second overload)
	method0, err := FindMethod(parsed, "foo0")
	if err != nil {
		t.Fatalf("unexpected error looking up 'foo0': %v", err)
	}
	if method0.RawName != "foo" {
		t.Errorf("expected RawName 'foo', got %q", method0.RawName)
	}

	// The two resolved methods should have different input counts
	if len(method.Inputs) == len(method0.Inputs) {
		t.Error("expected overloaded methods to have different input counts")
	}
}

func TestFindMethod_RawNameFallback(t *testing.T) {
	// Simulate a scenario where the map key differs from the RawName.
	// This happens when a method is the second overload (key="foo0", RawName="foo").
	parsed := ethabi.ABI{
		Methods: map[string]ethabi.Method{
			"foo0": {Name: "foo0", RawName: "foo"},
		},
	}

	method, err := FindMethod(parsed, "foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method.Name != "foo0" {
		t.Errorf("expected Name 'foo0', got %q", method.Name)
	}
}

func TestListMethods_ERC20(t *testing.T) {
	parsed := loadERC20(t)

	methods := ListMethods(parsed)
	if len(methods) == 0 {
		t.Fatal("expected non-empty methods list")
	}

	// Verify the list is sorted
	for i := 1; i < len(methods); i++ {
		if methods[i] < methods[i-1] {
			t.Errorf("methods not sorted: %q comes after %q", methods[i], methods[i-1])
		}
	}

	// Verify known methods are present (as signatures)
	joined := strings.Join(methods, ",")
	for _, name := range []string{"transfer", "approve", "balanceOf"} {
		if !strings.Contains(joined, name) {
			t.Errorf("expected method %q in list, got: %v", name, methods)
		}
	}
}

func TestListMethods_EmptyABI(t *testing.T) {
	methods := ListMethods(ethabi.ABI{})
	if len(methods) != 0 {
		t.Errorf("expected empty list for empty ABI, got: %v", methods)
	}
}
