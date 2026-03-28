package converter

import (
	"testing"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

func TestConvertArgs_ReturnsNotImplemented(t *testing.T) {
	_, err := ConvertArgs([]string{"1"}, ethabi.Arguments{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "not implemented" {
		t.Fatalf("expected 'not implemented', got %q", err.Error())
	}
}

func TestConvertArg_ReturnsNotImplemented(t *testing.T) {
	_, err := ConvertArg("1", ethabi.Type{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "not implemented" {
		t.Fatalf("expected 'not implemented', got %q", err.Error())
	}
}
