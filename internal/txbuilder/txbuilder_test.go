package txbuilder

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestBuild_DefaultParams(t *testing.T) {
	calldata := []byte{0xa9, 0x05, 0x9c, 0xbb}
	to := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	result, err := Build(calldata, to, TxParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must start with 0x prefix
	if !strings.HasPrefix(result, "0x") {
		t.Fatalf("expected 0x prefix, got %q", result)
	}

	// Must start with 0x02 (EIP-2718 DynamicFeeTx type prefix)
	if !strings.HasPrefix(result, "0x02") {
		t.Fatalf("expected 0x02 type prefix, got %q", result[:6])
	}

	// Deserialize and verify defaults were applied
	raw, err := hex.DecodeString(result[2:])
	if err != nil {
		t.Fatalf("hex decode failed: %v", err)
	}

	var tx types.Transaction
	if err := tx.UnmarshalBinary(raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Default ChainID should be 1
	if tx.ChainId().Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("expected chain ID 1, got %s", tx.ChainId())
	}

	// Default nonce should be 0
	if tx.Nonce() != 0 {
		t.Fatalf("expected nonce 0, got %d", tx.Nonce())
	}

	// Default GasTipCap should be 0
	if tx.GasTipCap().Cmp(big.NewInt(0)) != 0 {
		t.Fatalf("expected gas tip cap 0, got %s", tx.GasTipCap())
	}

	// Default GasFeeCap should be 0
	if tx.GasFeeCap().Cmp(big.NewInt(0)) != 0 {
		t.Fatalf("expected gas fee cap 0, got %s", tx.GasFeeCap())
	}

	// Default Value should be 0
	if tx.Value().Cmp(big.NewInt(0)) != 0 {
		t.Fatalf("expected value 0, got %s", tx.Value())
	}

	// Default GasLimit should be 0
	if tx.Gas() != 0 {
		t.Fatalf("expected gas limit 0, got %d", tx.Gas())
	}

	// To address should match
	if *tx.To() != to {
		t.Fatalf("expected to %s, got %s", to.Hex(), tx.To().Hex())
	}

	// Data should match
	if !equalBytes(tx.Data(), calldata) {
		t.Fatalf("expected data %x, got %x", calldata, tx.Data())
	}
}

func TestBuild_CustomChainID(t *testing.T) {
	calldata := []byte{0x01}
	to := common.HexToAddress("0xdead000000000000000000000000000000000000")

	result, err := Build(calldata, to, TxParams{
		ChainID: big.NewInt(137),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tx := deserializeTx(t, result)

	if tx.ChainId().Cmp(big.NewInt(137)) != 0 {
		t.Fatalf("expected chain ID 137, got %s", tx.ChainId())
	}
}

func TestBuild_CustomValue1ETH(t *testing.T) {
	calldata := []byte{0x01}
	to := common.HexToAddress("0xdead000000000000000000000000000000000000")

	oneETH := new(big.Int)
	oneETH.SetString("1000000000000000000", 10) // 1 ETH in wei

	result, err := Build(calldata, to, TxParams{
		Value: oneETH,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tx := deserializeTx(t, result)

	if tx.Value().Cmp(oneETH) != 0 {
		t.Fatalf("expected value %s, got %s", oneETH, tx.Value())
	}

	// ChainID should still default to 1
	if tx.ChainId().Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("expected default chain ID 1, got %s", tx.ChainId())
	}
}

func TestBuild_NilParams(t *testing.T) {
	calldata := []byte{0x01}
	to := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	// All nil/zero params — should not panic
	result, err := Build(calldata, to, TxParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(result, "0x02") {
		t.Fatalf("expected 0x02 prefix, got %q", result)
	}
}

func TestBuild_EmptyCalldata(t *testing.T) {
	to := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	result, err := Build([]byte{}, to, TxParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tx := deserializeTx(t, result)

	if len(tx.Data()) != 0 {
		t.Fatalf("expected empty data, got %x", tx.Data())
	}

	// Should still produce valid output with 0x02 prefix
	if !strings.HasPrefix(result, "0x02") {
		t.Fatalf("expected 0x02 prefix, got %q", result)
	}
}

func TestBuild_NilCalldata(t *testing.T) {
	to := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	result, err := Build(nil, to, TxParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(result, "0x02") {
		t.Fatalf("expected 0x02 prefix, got %q", result)
	}
}

func TestBuild_AllCustomParams(t *testing.T) {
	calldata := []byte{0xde, 0xad, 0xbe, 0xef}
	to := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")

	params := TxParams{
		ChainID:   big.NewInt(42161), // Arbitrum
		Nonce:     42,
		GasTipCap: big.NewInt(1000000000),  // 1 gwei
		GasFeeCap: big.NewInt(50000000000), // 50 gwei
		GasLimit:  21000,
		Value:     big.NewInt(500000000000000000), // 0.5 ETH
	}

	result, err := Build(calldata, to, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tx := deserializeTx(t, result)

	if tx.ChainId().Cmp(big.NewInt(42161)) != 0 {
		t.Fatalf("expected chain ID 42161, got %s", tx.ChainId())
	}
	if tx.Nonce() != 42 {
		t.Fatalf("expected nonce 42, got %d", tx.Nonce())
	}
	if tx.GasTipCap().Cmp(big.NewInt(1000000000)) != 0 {
		t.Fatalf("expected gas tip cap 1000000000, got %s", tx.GasTipCap())
	}
	if tx.GasFeeCap().Cmp(big.NewInt(50000000000)) != 0 {
		t.Fatalf("expected gas fee cap 50000000000, got %s", tx.GasFeeCap())
	}
	if tx.Gas() != 21000 {
		t.Fatalf("expected gas limit 21000, got %d", tx.Gas())
	}
	if tx.Value().Cmp(big.NewInt(500000000000000000)) != 0 {
		t.Fatalf("expected value 500000000000000000, got %s", tx.Value())
	}
	if !equalBytes(tx.Data(), calldata) {
		t.Fatalf("expected data %x, got %x", calldata, tx.Data())
	}
	if *tx.To() != to {
		t.Fatalf("expected to %s, got %s", to.Hex(), tx.To().Hex())
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

// Helper: deserialize hex string to transaction
func deserializeTx(t *testing.T, hexStr string) *types.Transaction {
	t.Helper()

	if !strings.HasPrefix(hexStr, "0x") {
		t.Fatalf("expected 0x prefix, got %q", hexStr)
	}

	raw, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		t.Fatalf("hex decode failed: %v", err)
	}

	var tx types.Transaction
	if err := tx.UnmarshalBinary(raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	return &tx
}

// Helper: compare byte slices
func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
