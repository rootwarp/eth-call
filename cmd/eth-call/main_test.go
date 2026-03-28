// Package main is the entrypoint for the eth-call CLI tool.
package main

import (
	"strings"
	"testing"
)

func TestBuildApp_ReturnsApp(t *testing.T) {
	app := buildApp()
	if app == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestBuildApp_Name(t *testing.T) {
	app := buildApp()
	if app.Name != "eth-call" {
		t.Fatalf("expected name 'eth-call', got %q", app.Name)
	}
}

func TestBuildApp_Version(t *testing.T) {
	app := buildApp()
	if app.Version != "0.1.0" {
		t.Fatalf("expected version '0.1.0', got %q", app.Version)
	}
}

func TestBuildApp_HasRequiredFlags(t *testing.T) {
	app := buildApp()

	flagNames := make(map[string]bool)
	for _, f := range app.Flags {
		for _, name := range f.Names() {
			flagNames[name] = true
		}
	}

	required := []string{"abi", "to", "chain-id", "value", "calldata-only", "rpc"}
	for _, name := range required {
		if !flagNames[name] {
			t.Errorf("missing flag: --%s", name)
		}
	}
}

func TestBuildApp_HelpRuns(t *testing.T) {
	app := buildApp()
	err := app.Run([]string{"eth-call", "--help"})
	if err != nil {
		t.Fatalf("--help returned error: %v", err)
	}
}

func TestBuildApp_ActionReturnsNotImplemented(t *testing.T) {
	app := buildApp()
	err := app.Run([]string{"eth-call", "--abi", "test.json", "--to", "0x0000000000000000000000000000000000000000", "transfer"})
	if err == nil {
		t.Fatal("expected error from stub action")
	}
	if err.Error() != "not implemented" {
		t.Fatalf("expected 'not implemented', got %q", err.Error())
	}
}

func TestBuildApp_BeforeHook_ValidAddress(t *testing.T) {
	app := buildApp()
	err := app.Run([]string{"eth-call", "--abi", "test.json", "--to", "0x0000000000000000000000000000000000000000", "transfer"})
	// Should pass Before hook and reach Action (which returns "not implemented")
	if err == nil {
		t.Fatal("expected error from stub action")
	}
	if err.Error() != "not implemented" {
		t.Fatalf("expected 'not implemented' after valid address, got %q", err.Error())
	}
}

func TestBuildApp_BeforeHook_InvalidAddress(t *testing.T) {
	app := buildApp()
	err := app.Run([]string{"eth-call", "--abi", "test.json", "--to", "not-an-address", "transfer"})
	if err == nil {
		t.Fatal("expected error for invalid address")
	}
	expected := "invalid address: not-an-address (expected 0x-prefixed 40-character hex)"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestBuildApp_BeforeHook_InvalidAddress_NoPrefixShort(t *testing.T) {
	app := buildApp()
	err := app.Run([]string{"eth-call", "--abi", "test.json", "--to", "0x123", "transfer"})
	if err == nil {
		t.Fatal("expected error for short address")
	}
	if !strings.Contains(err.Error(), "invalid address: 0x123") {
		t.Fatalf("expected invalid address error, got %q", err.Error())
	}
}

func TestBuildApp_BeforeHook_HelpSkipsValidation(t *testing.T) {
	app := buildApp()
	// --help should not trigger address validation
	err := app.Run([]string{"eth-call", "--help"})
	if err != nil {
		t.Fatalf("--help should not return error, got: %v", err)
	}
}

func TestBuildApp_Description_HasExamples(t *testing.T) {
	app := buildApp()
	if app.Description == "" {
		t.Fatal("expected non-empty Description with usage examples")
	}
	if !strings.Contains(app.Description, "transfer") {
		t.Error("Description should include transfer example")
	}
	if !strings.Contains(app.Description, "balanceOf") {
		t.Error("Description should include balanceOf example")
	}
	if !strings.Contains(app.Description, "--calldata-only") {
		t.Error("Description should include --calldata-only example")
	}
}
