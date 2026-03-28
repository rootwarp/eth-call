package encoder

import (
	"testing"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

func TestEncode_ReturnsNotImplemented(t *testing.T) {
	_, err := Encode(ethabi.ABI{}, "transfer", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "not implemented" {
		t.Fatalf("expected 'not implemented', got %q", err.Error())
	}
}

func TestEncodeToHex_ReturnsNotImplemented(t *testing.T) {
	_, err := EncodeToHex(ethabi.ABI{}, "transfer", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "not implemented" {
		t.Fatalf("expected 'not implemented', got %q", err.Error())
	}
}
