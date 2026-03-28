// Package main is the entrypoint for the eth-call CLI tool.
package main

import (
	"bytes"
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

// --- Pipeline integration tests ---

func TestAction_ERC20Transfer(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(output, "0x02") {
		t.Fatalf("expected output starting with 0x02 (DynamicFeeTx), got %q", output)
	}
}

func TestAction_CalldataOnly(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--calldata-only",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	// transfer(address,uint256) selector = 0xa9059cbb
	if !strings.HasPrefix(output, "0xa9059cbb") {
		t.Fatalf("expected calldata starting with transfer selector 0xa9059cbb, got %q", output)
	}
}

func TestAction_NoMethodListsMethods(t *testing.T) {
	app := buildApp()
	var stderr bytes.Buffer
	app.ErrWriter = &stderr

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stderr.String()
	if !strings.Contains(output, "Available methods:") {
		t.Fatalf("expected method listing, got %q", output)
	}
	if !strings.Contains(output, "transfer") {
		t.Fatalf("expected 'transfer' in method listing, got %q", output)
	}
}

func TestAction_ValueDecimal(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--value", "1000000000000000000",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(output, "0x02") {
		t.Fatalf("expected output starting with 0x02, got %q", output)
	}
}

func TestAction_ValueHex(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--value", "0xde0b6b3a7640000",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(output, "0x02") {
		t.Fatalf("expected output starting with 0x02, got %q", output)
	}
}

func TestAction_InvalidValue(t *testing.T) {
	app := buildApp()

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--value", "not-a-number",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
	})
	if err == nil {
		t.Fatal("expected error for invalid value")
	}
	if !strings.Contains(err.Error(), "invalid --value") {
		t.Fatalf("expected 'invalid --value' error, got %q", err.Error())
	}
}

func TestAction_ChainIDPassthrough(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--chain-id", "137",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(output, "0x02") {
		t.Fatalf("expected output starting with 0x02, got %q", output)
	}
}

func TestAction_EmptyCalldata(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	// Use a method with no arguments (totalSupply)
	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"totalSupply",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(output, "0x02") {
		t.Fatalf("expected output starting with 0x02, got %q", output)
	}
}
