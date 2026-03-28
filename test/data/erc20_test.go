package data

import (
	"os"
	"testing"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

func TestERC20Fixture_IsValidABI(t *testing.T) {
	f, err := os.Open("erc20.json")
	if err != nil {
		t.Fatalf("failed to open erc20.json: %v", err)
	}
	defer func() { _ = f.Close() }()

	parsedABI, err := ethabi.JSON(f)
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	requiredMethods := []string{"transfer", "approve", "balanceOf", "allowance", "totalSupply"}
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

func TestERC20Fixture_TransferInputs(t *testing.T) {
	f, err := os.Open("erc20.json")
	if err != nil {
		t.Fatalf("failed to open erc20.json: %v", err)
	}
	defer func() { _ = f.Close() }()

	parsedABI, err := ethabi.JSON(f)
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	transfer, ok := parsedABI.Methods["transfer"]
	if !ok {
		t.Fatal("transfer method not found")
	}

	if len(transfer.Inputs) != 2 {
		t.Fatalf("expected 2 inputs for transfer, got %d", len(transfer.Inputs))
	}

	if transfer.Inputs[0].Type.String() != "address" {
		t.Errorf("expected first input type 'address', got %q", transfer.Inputs[0].Type.String())
	}

	if transfer.Inputs[1].Type.String() != "uint256" {
		t.Errorf("expected second input type 'uint256', got %q", transfer.Inputs[1].Type.String())
	}
}
