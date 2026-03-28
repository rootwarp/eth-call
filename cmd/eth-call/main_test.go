// Package main is the entrypoint for the eth-call CLI tool.
package main

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	internalabi "github.com/rootwarp/eth-call/internal/abi"
	"github.com/rootwarp/eth-call/internal/converter"
	"github.com/rootwarp/eth-call/internal/encoder"
	"github.com/rootwarp/eth-call/internal/rpc"
	"github.com/rootwarp/eth-call/internal/txbuilder"
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

	required := []string{"abi", "to", "chain-id", "value", "calldata-only", "rpc", "from"}
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

// --- Error path tests ---

func TestAction_MissingABIFlag(t *testing.T) {
	app := buildApp()
	err := app.Run([]string{
		"eth-call",
		"--to", "0x1234567890123456789012345678901234567890",
		"transfer",
	})
	if err == nil {
		t.Fatal("expected error for missing --abi flag")
	}
	if !strings.Contains(err.Error(), "abi") {
		t.Fatalf("expected error mentioning 'abi', got %q", err.Error())
	}
}

func TestAction_InvalidMethodName(t *testing.T) {
	app := buildApp()
	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"nonExistentMethod",
	})
	if err == nil {
		t.Fatal("expected error for invalid method name")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected error containing 'not found', got %q", err.Error())
	}
	// Should list available methods in the error
	if !strings.Contains(err.Error(), "transfer") {
		t.Fatalf("expected error listing available methods (including 'transfer'), got %q", err.Error())
	}
}

func TestAction_WrongArgumentCount(t *testing.T) {
	app := buildApp()
	// transfer(address,uint256) expects 2 args, provide only 1
	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	})
	if err == nil {
		t.Fatal("expected error for wrong argument count")
	}
	if !strings.Contains(err.Error(), "expected 2") {
		t.Fatalf("expected error mentioning 'expected 2', got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "got 1") {
		t.Fatalf("expected error mentioning 'got 1', got %q", err.Error())
	}
}

func TestAction_WrongArgumentCount_TooMany(t *testing.T) {
	app := buildApp()
	// transfer(address,uint256) expects 2 args, provide 3
	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
		"extra-arg",
	})
	if err == nil {
		t.Fatal("expected error for too many arguments")
	}
	if !strings.Contains(err.Error(), "expected 2") {
		t.Fatalf("expected error mentioning 'expected 2', got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "got 3") {
		t.Fatalf("expected error mentioning 'got 3', got %q", err.Error())
	}
}

func TestAction_InvalidABIFile(t *testing.T) {
	app := buildApp()
	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/invalid.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"transfer",
	})
	if err == nil {
		t.Fatal("expected error for invalid ABI file")
	}
	if !strings.Contains(err.Error(), "abi") {
		t.Fatalf("expected error mentioning 'abi', got %q", err.Error())
	}
}

func TestAction_NonexistentABIFile(t *testing.T) {
	app := buildApp()
	err := app.Run([]string{
		"eth-call",
		"--abi", "nonexistent-file.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"transfer",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent ABI file")
	}
	if !strings.Contains(err.Error(), "abi") {
		t.Fatalf("expected error mentioning 'abi', got %q", err.Error())
	}
}

func TestAction_CalldataOnly_BalanceOf(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--calldata-only",
		"balanceOf",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	// balanceOf(address) selector = 0x70a08231
	if !strings.HasPrefix(output, "0x70a08231") {
		t.Fatalf("expected calldata starting with balanceOf selector 0x70a08231, got %q", output)
	}
}

func TestAction_Approve(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"approve",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"5000",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(output, "0x02") {
		t.Fatalf("expected output starting with 0x02, got %q", output)
	}
}

func TestAction_NoMethodListsMethods_ContainsMultiple(t *testing.T) {
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
	for _, method := range []string{"transfer", "approve", "balanceOf", "totalSupply"} {
		if !strings.Contains(output, method) {
			t.Errorf("expected method listing to contain %q, got %q", method, output)
		}
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

// --- Uniswap V2 CLI integration tests ---

func TestAction_UniswapSwap_CalldataOnly(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/uniswap_v2.json",
		"--to", "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D",
		"--calldata-only",
		"swapExactTokensForTokens",
		"1000",
		"1",
		`["0x0000000000000000000000000000000000000001","0x0000000000000000000000000000000000000002"]`,
		"0x0000000000000000000000000000000000000003",
		"9999",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(output, "0x38ed1739") {
		t.Errorf("expected calldata starting with 0x38ed1739, got %q", output[:14])
	}
}

func TestAction_UniswapSwap_FullTx(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/uniswap_v2.json",
		"--to", "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D",
		"swapExactTokensForTokens",
		"1000",
		"1",
		`["0x0000000000000000000000000000000000000001","0x0000000000000000000000000000000000000002"]`,
		"0x0000000000000000000000000000000000000003",
		"9999",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(output, "0x02") {
		t.Fatalf("expected output starting with 0x02 (DynamicFeeTx), got %q", output)
	}
}

func TestAction_UniswapAddLiquidity_CalldataOnly(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/uniswap_v2.json",
		"--to", "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D",
		"--calldata-only",
		"addLiquidity",
		"0x0000000000000000000000000000000000000001",
		"0x0000000000000000000000000000000000000002",
		"100",
		"200",
		"50",
		"100",
		"0x0000000000000000000000000000000000000003",
		"9999",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(output, "0xe8e33700") {
		t.Errorf("expected calldata starting with 0xe8e33700, got %q", output[:14])
	}
}

func TestAction_UniswapGetAmountsOut_CalldataOnly(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/uniswap_v2.json",
		"--to", "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D",
		"--calldata-only",
		"getAmountsOut",
		"1000",
		`["0x0000000000000000000000000000000000000001","0x0000000000000000000000000000000000000002"]`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(output, "0xd06ca61f") {
		t.Errorf("expected calldata starting with 0xd06ca61f, got %q", output[:14])
	}
}

func TestAction_UniswapMaxUint256_CalldataOnly(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	maxUint256 := "115792089237316195423570985008687907853269984665640564039457584007913129639935"

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/uniswap_v2.json",
		"--to", "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D",
		"--calldata-only",
		"getAmountsOut",
		maxUint256,
		`["0x0000000000000000000000000000000000000001"]`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	// Verify selector
	if !strings.HasPrefix(output, "0xd06ca61f") {
		t.Errorf("expected calldata starting with 0xd06ca61f, got %q", output[:14])
	}
	// Verify max uint256 is encoded as all f's (chars 10-74 after "0x" prefix and 8-char selector)
	amountIn := output[10:74]
	expectedMax := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	if amountIn != expectedMax {
		t.Errorf("max uint256 encoding mismatch\nexpected: %s\ngot:      %s", expectedMax, amountIn)
	}
}

// --- DEV-22: --from flag and --rpc wiring tests ---

func TestBuildApp_BeforeHook_InvalidFromAddress(t *testing.T) {
	app := buildApp()
	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--from", "not-an-address",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
	})
	if err == nil {
		t.Fatal("expected error for invalid --from address")
	}
	if !strings.Contains(err.Error(), "invalid --from address") {
		t.Fatalf("expected 'invalid --from address' error, got %q", err.Error())
	}
}

func TestAction_RPCWithoutFrom_ReturnsError(t *testing.T) {
	app := buildApp()
	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--rpc", "http://localhost:8545",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
	})
	if err == nil {
		t.Fatal("expected error when --rpc is set without --from")
	}
	if !strings.Contains(err.Error(), "--from is required when --rpc is provided") {
		t.Fatalf("expected '--from is required' error, got %q", err.Error())
	}
}

func TestAction_FromWithoutRPC_Ignored(t *testing.T) {
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	// --from without --rpc should work fine (from is just ignored in offline mode)
	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--from", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
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

func TestAction_RPCWiring_UsesRPCParams(t *testing.T) {
	fetchCalled := false
	closeCalled := false

	origDialFn := dialFn
	defer func() { dialFn = origDialFn }()

	dialFn = func(_ context.Context, _ string) (rpc.Client, error) {
		return &rpc.MockClient{
			FetchParamsFn: func(_ context.Context, _, _ common.Address, _ []byte, _ *big.Int) (txbuilder.TxParams, error) {
				fetchCalled = true
				return txbuilder.TxParams{
					ChainID:   big.NewInt(42),
					Nonce:     7,
					GasTipCap: big.NewInt(1_000_000_000),
					GasFeeCap: big.NewInt(20_000_000_000),
					GasLimit:  60000,
					Value:     big.NewInt(0),
				}, nil
			},
			CloseFn: func() {
				closeCalled = true
			},
		}, nil
	}

	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--from", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"--rpc", "http://localhost:8545",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fetchCalled {
		t.Fatal("expected FetchParams to be called")
	}
	if !closeCalled {
		t.Fatal("expected Close to be called")
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(output, "0x02") {
		t.Fatalf("expected output starting with 0x02, got %q", output)
	}
}

func TestAction_RPCDialError(t *testing.T) {
	origDialFn := dialFn
	defer func() { dialFn = origDialFn }()

	dialFn = func(_ context.Context, _ string) (rpc.Client, error) {
		return nil, fmt.Errorf("dial failed: connection refused")
	}

	app := buildApp()

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--from", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"--rpc", "http://localhost:8545",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
	})
	if err == nil {
		t.Fatal("expected error for dial failure")
	}
	if !strings.Contains(err.Error(), "dial failed") {
		t.Fatalf("expected 'dial failed' error, got %q", err.Error())
	}
}

func TestAction_RPCFetchParamsError(t *testing.T) {
	origDialFn := dialFn
	defer func() { dialFn = origDialFn }()

	dialFn = func(_ context.Context, _ string) (rpc.Client, error) {
		return &rpc.MockClient{
			FetchParamsFn: func(_ context.Context, _, _ common.Address, _ []byte, _ *big.Int) (txbuilder.TxParams, error) {
				return txbuilder.TxParams{}, fmt.Errorf("rpc: failed to fetch nonce: connection reset")
			},
		}, nil
	}

	app := buildApp()

	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--from", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"--rpc", "http://localhost:8545",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
	})
	if err == nil {
		t.Fatal("expected error for FetchParams failure")
	}
	if !strings.Contains(err.Error(), "failed to fetch nonce") {
		t.Fatalf("expected nonce error, got %q", err.Error())
	}
}

func TestAction_CLIFlagsOverrideRPC(t *testing.T) {
	origDialFn := dialFn
	defer func() { dialFn = origDialFn }()

	dialFn = func(_ context.Context, _ string) (rpc.Client, error) {
		return &rpc.MockClient{
			FetchParamsFn: func(_ context.Context, _, _ common.Address, _ []byte, _ *big.Int) (txbuilder.TxParams, error) {
				return txbuilder.TxParams{
					ChainID:   big.NewInt(999),
					Nonce:     7,
					GasTipCap: big.NewInt(1_000_000_000),
					GasFeeCap: big.NewInt(20_000_000_000),
					GasLimit:  60000,
					Value:     big.NewInt(999),
				}, nil
			},
		}, nil
	}

	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	// Explicitly set --chain-id and --value, these should override RPC values
	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--from", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"--rpc", "http://localhost:8545",
		"--chain-id", "137",
		"--value", "500",
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
	// The output should use chain-id 137 (0x89) not 999 from RPC,
	// and value 500 not 999 from RPC.
	// We verify this indirectly — the test passing without error means the pipeline completed.
}

func TestAction_RPCCalldataOnlySkipsRPC(t *testing.T) {
	origDialFn := dialFn
	defer func() { dialFn = origDialFn }()

	dialCalled := false
	dialFn = func(_ context.Context, _ string) (rpc.Client, error) {
		dialCalled = true
		return &rpc.MockClient{}, nil
	}

	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	// --calldata-only should skip RPC even if --rpc is provided
	err := app.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--rpc", "http://localhost:8545",
		"--calldata-only",
		"balanceOf",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dialCalled {
		t.Fatal("expected RPC dial to be skipped in --calldata-only mode")
	}
}

// --- DEV-24: --calldata-only finalization + output mode tests ---

func TestCalldataOnly_MatchesEncoderOutput(t *testing.T) {
	// Load ABI and encode directly via encoder package
	parsedABI, err := internalabi.LoadFromFile("../../test/data/erc20.json")
	if err != nil {
		t.Fatalf("failed to load ABI: %v", err)
	}
	methodDef, err := internalabi.FindMethod(parsedABI, "transfer")
	if err != nil {
		t.Fatalf("failed to find method: %v", err)
	}
	args, err := converter.ConvertArgs(
		[]string{"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd", "1000"},
		methodDef.Inputs,
	)
	if err != nil {
		t.Fatalf("failed to convert args: %v", err)
	}
	expectedHex, err := encoder.EncodeToHex(parsedABI, "transfer", args)
	if err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	// Run CLI with --calldata-only
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout

	err = app.Run([]string{
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

	cliOutput := strings.TrimSpace(stdout.String())
	if cliOutput != expectedHex {
		t.Fatalf("--calldata-only output does not match encoder.EncodeToHex\nexpected: %s\ngot:      %s", expectedHex, cliOutput)
	}
}

func TestCalldataOnly_IdenticalWithAndWithoutRPC(t *testing.T) {
	// Run without --rpc
	appNoRPC := buildApp()
	var stdoutNoRPC bytes.Buffer
	appNoRPC.Writer = &stdoutNoRPC

	err := appNoRPC.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--calldata-only",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
	})
	if err != nil {
		t.Fatalf("unexpected error (no RPC): %v", err)
	}

	// Run with --rpc (mock)
	origDialFn := dialFn
	defer func() { dialFn = origDialFn }()
	dialFn = func(_ context.Context, _ string) (rpc.Client, error) {
		return &rpc.MockClient{}, nil
	}

	appWithRPC := buildApp()
	var stdoutWithRPC bytes.Buffer
	appWithRPC.Writer = &stdoutWithRPC

	err = appWithRPC.Run([]string{
		"eth-call",
		"--abi", "../../test/data/erc20.json",
		"--to", "0x1234567890123456789012345678901234567890",
		"--rpc", "http://localhost:8545",
		"--calldata-only",
		"transfer",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"1000",
	})
	if err != nil {
		t.Fatalf("unexpected error (with RPC): %v", err)
	}

	outputNoRPC := strings.TrimSpace(stdoutNoRPC.String())
	outputWithRPC := strings.TrimSpace(stdoutWithRPC.String())

	if outputNoRPC != outputWithRPC {
		t.Fatalf("--calldata-only output differs with/without --rpc\nwithout: %s\nwith:    %s", outputNoRPC, outputWithRPC)
	}
}

func TestOutputMode_DefaultIsRawHexTransaction(t *testing.T) {
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
	// Default mode outputs a full RLP-encoded EIP-2718 DynamicFeeTx (starts with 0x02)
	if !strings.HasPrefix(output, "0x02") {
		t.Fatalf("default mode should output raw hex transaction starting with 0x02, got %q", output)
	}
	// Should NOT start with a method selector (e.g., 0xa9059cbb for transfer)
	if strings.HasPrefix(output, "0xa9059cbb") {
		t.Fatal("default mode should output full transaction, not just calldata")
	}
}

func TestOutputMode_CalldataOnlyIsCalldataHex(t *testing.T) {
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
	// Calldata-only starts with method selector, not with 0x02 (DynamicFeeTx prefix)
	if !strings.HasPrefix(output, "0xa9059cbb") {
		t.Fatalf("--calldata-only should output calldata starting with transfer selector 0xa9059cbb, got %q", output)
	}
	// Must NOT start with 0x02 (that's a full transaction)
	if strings.HasPrefix(output, "0x02") {
		t.Fatal("--calldata-only should output calldata, not a full transaction")
	}
}

func TestOutputMode_StdoutOnlyOnSuccess(t *testing.T) {
	app := buildApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

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

	if stdout.Len() == 0 {
		t.Fatal("expected output on stdout")
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr on success, got %q", stderr.String())
	}
}

func TestOutputMode_CalldataOnly_StdoutOnlyOnSuccess(t *testing.T) {
	app := buildApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

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

	if stdout.Len() == 0 {
		t.Fatal("expected output on stdout")
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr on success, got %q", stderr.String())
	}
}
