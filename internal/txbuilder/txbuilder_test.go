package txbuilder

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestBuild_ReturnsNotImplemented(t *testing.T) {
	_, err := Build([]byte{0x01}, common.Address{}, TxParams{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "not implemented" {
		t.Fatalf("expected 'not implemented', got %q", err.Error())
	}
}

func TestBuildTx_ReturnsNotImplemented(t *testing.T) {
	_, err := BuildTx([]byte{0x01}, common.Address{}, TxParams{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "not implemented" {
		t.Fatalf("expected 'not implemented', got %q", err.Error())
	}
}

func TestTxParams_ZeroValue(t *testing.T) {
	params := TxParams{}
	if params.ChainID != nil {
		t.Fatal("expected nil ChainID")
	}
	if params.Nonce != 0 {
		t.Fatal("expected zero Nonce")
	}
	if params.GasTipCap != nil {
		t.Fatal("expected nil GasTipCap")
	}
	if params.GasFeeCap != nil {
		t.Fatal("expected nil GasFeeCap")
	}
	if params.GasLimit != 0 {
		t.Fatal("expected zero GasLimit")
	}
	if params.Value != nil {
		t.Fatal("expected nil Value")
	}
}
