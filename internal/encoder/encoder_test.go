package encoder

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func loadERC20ABI(t *testing.T) ethabi.ABI {
	t.Helper()
	parsed, err := ethabi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		t.Fatalf("failed to parse ERC-20 ABI: %v", err)
	}
	return parsed
}

func TestEncode_TransferSelector(t *testing.T) {
	parsed := loadERC20ABI(t)

	args := []interface{}{
		common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
		big.NewInt(1000),
	}

	calldata, err := Encode(parsed, "transfer", args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// transfer(address,uint256) selector = 0xa9059cbb
	gotSelector := hex.EncodeToString(calldata[:4])
	if gotSelector != "a9059cbb" {
		t.Errorf("expected selector a9059cbb, got %s", gotSelector)
	}
}

func TestEncode_ApproveSelector(t *testing.T) {
	parsed := loadERC20ABI(t)

	args := []interface{}{
		common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
		big.NewInt(500),
	}

	calldata, err := Encode(parsed, "approve", args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotSelector := hex.EncodeToString(calldata[:4])
	if gotSelector != "095ea7b3" {
		t.Errorf("expected selector 095ea7b3, got %s", gotSelector)
	}
}

func TestEncode_TransferFullCalldata(t *testing.T) {
	parsed := loadERC20ABI(t)

	recipient := common.HexToAddress("0x000000000000000000000000000000000000dEaD")
	amount := big.NewInt(1000)
	args := []interface{}{recipient, amount}

	calldata, err := Encode(parsed, "transfer", args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Known-good calldata for transfer(0x...dEaD, 1000):
	// selector: a9059cbb
	// address padded to 32 bytes
	// uint256 1000 = 0x3e8 padded to 32 bytes
	expected := "a9059cbb" +
		"000000000000000000000000000000000000000000000000000000000000dead" +
		"00000000000000000000000000000000000000000000000000000000000003e8"

	got := hex.EncodeToString(calldata)
	if got != expected {
		t.Errorf("calldata mismatch\nexpected: %s\ngot:      %s", expected, got)
	}
}

func TestEncode_BalanceOfSelector(t *testing.T) {
	parsed := loadERC20ABI(t)

	args := []interface{}{
		common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
	}

	calldata, err := Encode(parsed, "balanceOf", args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotSelector := hex.EncodeToString(calldata[:4])
	if gotSelector != "70a08231" {
		t.Errorf("expected selector 70a08231, got %s", gotSelector)
	}
}

func TestEncode_WrongArgCount(t *testing.T) {
	parsed := loadERC20ABI(t)

	// transfer expects 2 args, give 1
	args := []interface{}{
		common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
	}

	_, err := Encode(parsed, "transfer", args)
	if err == nil {
		t.Fatal("expected error for wrong arg count, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "encode:") {
		t.Errorf("error should contain 'encode:' prefix, got: %v", errMsg)
	}
	if !strings.Contains(errMsg, "transfer") {
		t.Errorf("error should mention method name, got: %v", errMsg)
	}
}

func TestEncode_NoArgs(t *testing.T) {
	parsed := loadERC20ABI(t)

	// totalSupply() takes no arguments
	calldata, err := Encode(parsed, "totalSupply", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be just the 4-byte selector
	if len(calldata) != 4 {
		t.Errorf("expected 4 bytes for no-arg method, got %d", len(calldata))
	}
}

func TestEncode_NonexistentMethod(t *testing.T) {
	parsed := loadERC20ABI(t)

	_, err := Encode(parsed, "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent method, got nil")
	}

	if !strings.Contains(err.Error(), "encode:") {
		t.Errorf("error should contain 'encode:' prefix, got: %v", err)
	}
}

func TestEncodeToHex_Prefix(t *testing.T) {
	parsed := loadERC20ABI(t)

	args := []interface{}{
		common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
		big.NewInt(1000),
	}

	hexStr, err := EncodeToHex(parsed, "transfer", args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(hexStr, "0x") {
		t.Errorf("expected 0x prefix, got: %s", hexStr)
	}
}

func TestEncodeToHex_TransferSelector(t *testing.T) {
	parsed := loadERC20ABI(t)

	args := []interface{}{
		common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
		big.NewInt(1000),
	}

	hexStr, err := EncodeToHex(parsed, "transfer", args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(hexStr, "0xa9059cbb") {
		t.Errorf("expected hex to start with 0xa9059cbb, got: %s", hexStr[:12])
	}
}

func TestEncodeToHex_ApproveSelector(t *testing.T) {
	parsed := loadERC20ABI(t)

	args := []interface{}{
		common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
		big.NewInt(500),
	}

	hexStr, err := EncodeToHex(parsed, "approve", args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(hexStr, "0x095ea7b3") {
		t.Errorf("expected hex to start with 0x095ea7b3, got: %s", hexStr[:12])
	}
}

func TestEncodeToHex_FullCalldata(t *testing.T) {
	parsed := loadERC20ABI(t)

	recipient := common.HexToAddress("0x000000000000000000000000000000000000dEaD")
	amount := big.NewInt(1000)
	args := []interface{}{recipient, amount}

	hexStr, err := EncodeToHex(parsed, "transfer", args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "0x" +
		"a9059cbb" +
		"000000000000000000000000000000000000000000000000000000000000dead" +
		"00000000000000000000000000000000000000000000000000000000000003e8"

	if hexStr != expected {
		t.Errorf("hex mismatch\nexpected: %s\ngot:      %s", expected, hexStr)
	}
}

func TestEncodeToHex_Error(t *testing.T) {
	parsed := loadERC20ABI(t)

	// Wrong arg count
	args := []interface{}{
		common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
	}

	_, err := EncodeToHex(parsed, "transfer", args)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Inline ERC-20 ABI for tests (avoids file-path dependency)
const erc20ABI = `[
  {"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"type":"function","stateMutability":"view"},
  {"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"type":"function","stateMutability":"view"},
  {"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"type":"function","stateMutability":"view"},
  {"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"type":"function","stateMutability":"view"},
  {"constant":true,"inputs":[{"name":"account","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"type":"function","stateMutability":"view"},
  {"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function","stateMutability":"nonpayable"},
  {"constant":true,"inputs":[{"name":"owner","type":"address"},{"name":"spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"type":"function","stateMutability":"view"},
  {"constant":false,"inputs":[{"name":"spender","type":"address"},{"name":"value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"type":"function","stateMutability":"nonpayable"},
  {"constant":false,"inputs":[{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"type":"function","stateMutability":"nonpayable"}
]`
