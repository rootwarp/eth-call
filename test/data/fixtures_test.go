package data

import (
	"os"
	"testing"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

func TestUniswapV2Fixture_IsValidABI(t *testing.T) {
	f, err := os.Open("uniswap_v2.json")
	if err != nil {
		t.Fatalf("failed to open uniswap_v2.json: %v", err)
	}
	defer func() { _ = f.Close() }()

	parsedABI, err := ethabi.JSON(f)
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	requiredMethods := []string{
		"swapExactTokensForTokens",
		"addLiquidity",
		"removeLiquidity",
		"getAmountsOut",
	}
	for _, name := range requiredMethods {
		found := false
		for _, m := range parsedABI.Methods {
			if m.RawName == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing required method: %s", name)
		}
	}
}

func TestUniswapV2Fixture_SwapExactTokensForTokens(t *testing.T) {
	f, err := os.Open("uniswap_v2.json")
	if err != nil {
		t.Fatalf("failed to open uniswap_v2.json: %v", err)
	}
	defer func() { _ = f.Close() }()

	parsedABI, err := ethabi.JSON(f)
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	method, ok := parsedABI.Methods["swapExactTokensForTokens"]
	if !ok {
		t.Fatal("swapExactTokensForTokens not found")
	}

	// swapExactTokensForTokens(uint256,uint256,address[],address,uint256)
	if len(method.Inputs) != 5 {
		t.Fatalf("expected 5 inputs, got %d", len(method.Inputs))
	}

	expectedTypes := []string{"uint256", "uint256", "address[]", "address", "uint256"}
	for i, exp := range expectedTypes {
		if method.Inputs[i].Type.String() != exp {
			t.Errorf("input %d: expected %q, got %q", i, exp, method.Inputs[i].Type.String())
		}
	}
}

func TestUniswapV2Fixture_AddLiquidity(t *testing.T) {
	f, err := os.Open("uniswap_v2.json")
	if err != nil {
		t.Fatalf("failed to open uniswap_v2.json: %v", err)
	}
	defer func() { _ = f.Close() }()

	parsedABI, err := ethabi.JSON(f)
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	method, ok := parsedABI.Methods["addLiquidity"]
	if !ok {
		t.Fatal("addLiquidity not found")
	}

	// addLiquidity(address,address,uint256,uint256,uint256,uint256,address,uint256)
	if len(method.Inputs) != 8 {
		t.Fatalf("expected 8 inputs, got %d", len(method.Inputs))
	}
}

func TestUniswapV2Fixture_GetAmountsOut(t *testing.T) {
	f, err := os.Open("uniswap_v2.json")
	if err != nil {
		t.Fatalf("failed to open uniswap_v2.json: %v", err)
	}
	defer func() { _ = f.Close() }()

	parsedABI, err := ethabi.JSON(f)
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	method, ok := parsedABI.Methods["getAmountsOut"]
	if !ok {
		t.Fatal("getAmountsOut not found")
	}

	// getAmountsOut(uint256,address[])
	if len(method.Inputs) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(method.Inputs))
	}
	if method.Inputs[0].Type.String() != "uint256" {
		t.Errorf("expected uint256, got %q", method.Inputs[0].Type.String())
	}
	if method.Inputs[1].Type.String() != "address[]" {
		t.Errorf("expected address[], got %q", method.Inputs[1].Type.String())
	}
}

func TestComplexFixture_IsValidABI(t *testing.T) {
	f, err := os.Open("complex.json")
	if err != nil {
		t.Fatalf("failed to open complex.json: %v", err)
	}
	defer func() { _ = f.Close() }()

	parsedABI, err := ethabi.JSON(f)
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	requiredMethods := []string{
		// Original methods
		"processTuple",
		"batchTransfer",
		"swapPair",
		"processNested",
		"storeHash",
		// New methods for full converter coverage
		"mixedInts",
		"dynamicBytesArray",
		"nestedArrays",
		"complexTuple",
	}
	for _, name := range requiredMethods {
		found := false
		for _, m := range parsedABI.Methods {
			if m.RawName == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing required method: %s", name)
		}
	}
}

func TestComplexFixture_MixedInts(t *testing.T) {
	f, err := os.Open("complex.json")
	if err != nil {
		t.Fatalf("failed to open complex.json: %v", err)
	}
	defer func() { _ = f.Close() }()

	parsedABI, err := ethabi.JSON(f)
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	method, ok := parsedABI.Methods["mixedInts"]
	if !ok {
		t.Fatal("mixedInts not found")
	}

	expectedTypes := []string{"uint8", "uint32", "uint128", "uint256", "int8", "int256"}
	if len(method.Inputs) != len(expectedTypes) {
		t.Fatalf("expected %d inputs, got %d", len(expectedTypes), len(method.Inputs))
	}
	for i, exp := range expectedTypes {
		if method.Inputs[i].Type.String() != exp {
			t.Errorf("input %d: expected %q, got %q", i, exp, method.Inputs[i].Type.String())
		}
	}
}

func TestComplexFixture_ComplexTuple(t *testing.T) {
	f, err := os.Open("complex.json")
	if err != nil {
		t.Fatalf("failed to open complex.json: %v", err)
	}
	defer func() { _ = f.Close() }()

	parsedABI, err := ethabi.JSON(f)
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	method, ok := parsedABI.Methods["complexTuple"]
	if !ok {
		t.Fatal("complexTuple not found")
	}

	// Should have a tuple with mixed types including nested tuple, arrays, bytes32
	if len(method.Inputs) != 1 {
		t.Fatalf("expected 1 tuple input, got %d", len(method.Inputs))
	}
	if method.Inputs[0].Type.String() != "(uint256,address,(bool,bytes32),uint256[])" {
		t.Errorf("unexpected tuple type: %q", method.Inputs[0].Type.String())
	}
}
