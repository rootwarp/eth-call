package converter

import (
	"math/big"
	"strings"
	"testing"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// makeType is a test helper that creates an ethabi.Type with the given T value.
func makeType(t byte) ethabi.Type {
	return ethabi.Type{T: t}
}

// makeIntType creates an ethabi.Type for integer types with a given size.
func makeIntType(t byte, size int) ethabi.Type {
	return ethabi.Type{T: t, Size: size}
}

func TestConvertArgs_WrongArgCount(t *testing.T) {
	inputs := ethabi.Arguments{
		{Name: "a", Type: makeType(ethabi.StringTy)},
		{Name: "b", Type: makeType(ethabi.StringTy)},
	}
	_, err := ConvertArgs([]string{"one"}, inputs)
	if err == nil {
		t.Fatal("expected error for wrong arg count, got nil")
	}
	want := "converter: expected 2 arguments, got 1"
	if err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}

func TestConvertArgs_WrongArgCount_TooMany(t *testing.T) {
	inputs := ethabi.Arguments{
		{Name: "a", Type: makeType(ethabi.StringTy)},
	}
	_, err := ConvertArgs([]string{"one", "two"}, inputs)
	if err == nil {
		t.Fatal("expected error for wrong arg count, got nil")
	}
	want := "converter: expected 1 arguments, got 2"
	if err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}

func TestConvertArg_UnsupportedType(t *testing.T) {
	typ := makeType(ethabi.HashTy)
	_, err := ConvertArg("test", typ)
	if err == nil {
		t.Fatal("expected error for unsupported type, got nil")
	}
	if !strings.Contains(err.Error(), "converter: unsupported type:") {
		t.Fatalf("expected unsupported type error, got %q", err.Error())
	}
}

func TestConvertArg_StringPassthrough(t *testing.T) {
	typ := makeType(ethabi.StringTy)
	result, err := ConvertArg("hello world", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(string)
	if !ok {
		t.Fatalf("expected string, got %T", result)
	}
	if str != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", str)
	}
}

func TestConvertArg_StringEmpty(t *testing.T) {
	typ := makeType(ethabi.StringTy)
	result, err := ConvertArg("", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(string)
	if !ok {
		t.Fatalf("expected string, got %T", result)
	}
	if str != "" {
		t.Fatalf("expected empty string, got %q", str)
	}
}

func TestConvertArgs_ValidStringArgs(t *testing.T) {
	inputs := ethabi.Arguments{
		{Name: "name", Type: makeType(ethabi.StringTy)},
		{Name: "desc", Type: makeType(ethabi.StringTy)},
	}
	result, err := ConvertArgs([]string{"alice", "hello"}, inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0] != "alice" {
		t.Fatalf("expected %q, got %v", "alice", result[0])
	}
	if result[1] != "hello" {
		t.Fatalf("expected %q, got %v", "hello", result[1])
	}
}

func TestConvertArgs_WrapsErrorWithIndex(t *testing.T) {
	inputs := ethabi.Arguments{
		{Name: "ok", Type: makeType(ethabi.StringTy)},
		{Name: "bad", Type: makeType(ethabi.HashTy)},
	}
	_, err := ConvertArgs([]string{"hello", "test"}, inputs)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "converter: arg[1] (bad):") {
		t.Fatalf("expected error with arg index wrapper, got %q", err.Error())
	}
}

func TestConvertArgs_ZeroArgs(t *testing.T) {
	result, err := ConvertArgs([]string{}, ethabi.Arguments{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result))
	}
}

func TestConvertArg_NotImplementedTypes(t *testing.T) {
	tests := []struct {
		name string
		typ  byte
	}{
		{"BytesTy", ethabi.BytesTy},
		{"FixedBytesTy", ethabi.FixedBytesTy},
		{"SliceTy", ethabi.SliceTy},
		{"ArrayTy", ethabi.ArrayTy},
		{"TupleTy", ethabi.TupleTy},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := makeType(tt.typ)
			_, err := ConvertArg("test", typ)
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tt.name)
			}
			if !strings.Contains(err.Error(), "not implemented") {
				t.Fatalf("expected 'not implemented' error for %s, got %q", tt.name, err.Error())
			}
		})
	}
}

// --- Address tests ---

func TestConvertArg_Address(t *testing.T) {
	typ := makeType(ethabi.AddressTy)

	t.Run("valid", func(t *testing.T) {
		result, err := ConvertArg("0xdAC17F958D2ee523a2206206994597C13D831ec7", typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		addr, ok := result.(common.Address)
		if !ok {
			t.Fatalf("expected common.Address, got %T", result)
		}
		want := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
		if addr != want {
			t.Fatalf("expected %s, got %s", want.Hex(), addr.Hex())
		}
	})

	t.Run("zero_address", func(t *testing.T) {
		result, err := ConvertArg("0x0000000000000000000000000000000000000000", typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		addr, ok := result.(common.Address)
		if !ok {
			t.Fatalf("expected common.Address, got %T", result)
		}
		if addr != (common.Address{}) {
			t.Fatalf("expected zero address, got %s", addr.Hex())
		}
	})

	t.Run("invalid_hex", func(t *testing.T) {
		_, err := ConvertArg("0xZZZ", typ)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid address") {
			t.Fatalf("expected invalid address error, got %q", err.Error())
		}
	})

	t.Run("short_input", func(t *testing.T) {
		_, err := ConvertArg("0x1234", typ)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid address") {
			t.Fatalf("expected invalid address error, got %q", err.Error())
		}
	})
}

// --- Bool tests ---

func TestConvertArg_Bool(t *testing.T) {
	typ := makeType(ethabi.BoolTy)

	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{"true", "true", true, false},
		{"false", "false", false, false},
		{"one", "1", true, false},
		{"zero", "0", false, false},
		{"yes_invalid", "yes", false, true},
		{"two_invalid", "2", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertArg(tt.input, typ)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "invalid bool") {
					t.Fatalf("expected invalid bool error, got %q", err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			v, ok := result.(bool)
			if !ok {
				t.Fatalf("expected bool, got %T", result)
			}
			if v != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, v)
			}
		})
	}
}

// --- Uint tests ---

func TestConvertArg_Uint8(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    uint8
		wantErr bool
	}{
		{"zero", "0", 0, false},
		{"basic", "42", 42, false},
		{"max", "255", 255, false},
		{"hex", "0xff", 255, false},
		{"overflow", "256", 0, true},
		{"negative", "-1", 0, true},
		{"invalid", "abc", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := makeIntType(ethabi.UintTy, 8)
			result, err := ConvertArg(tt.input, typ)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result=%v)", result)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			v, ok := result.(uint8)
			if !ok {
				t.Fatalf("expected uint8, got %T", result)
			}
			if v != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, v)
			}
		})
	}
}

func TestConvertArg_Uint16(t *testing.T) {
	typ := makeIntType(ethabi.UintTy, 16)
	result, err := ConvertArg("65535", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := result.(uint16)
	if !ok {
		t.Fatalf("expected uint16, got %T", result)
	}
	if v != 65535 {
		t.Fatalf("expected 65535, got %d", v)
	}
}

func TestConvertArg_Uint32(t *testing.T) {
	typ := makeIntType(ethabi.UintTy, 32)
	result, err := ConvertArg("4294967295", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := result.(uint32)
	if !ok {
		t.Fatalf("expected uint32, got %T", result)
	}
	if v != 4294967295 {
		t.Fatalf("expected 4294967295, got %d", v)
	}
}

func TestConvertArg_Uint64(t *testing.T) {
	typ := makeIntType(ethabi.UintTy, 64)
	result, err := ConvertArg("18446744073709551615", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := result.(uint64)
	if !ok {
		t.Fatalf("expected uint64, got %T", result)
	}
	if v != 18446744073709551615 {
		t.Fatalf("expected max uint64, got %d", v)
	}
}

func TestConvertArg_Uint256(t *testing.T) {
	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

	tests := []struct {
		want    *big.Int
		name    string
		input   string
		wantErr bool
	}{
		{big.NewInt(0), "zero", "0", false},
		{big.NewInt(1000), "basic", "1000", false},
		{maxUint256, "max", maxUint256.String(), false},
		{big.NewInt(100), "hex", "0x64", false},
		{nil, "negative", "-1", true},
		{nil, "overflow", new(big.Int).Add(maxUint256, big.NewInt(1)).String(), true},
		{nil, "invalid", "xyz", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := makeIntType(ethabi.UintTy, 256)
			result, err := ConvertArg(tt.input, typ)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result=%v)", result)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			v, ok := result.(*big.Int)
			if !ok {
				t.Fatalf("expected *big.Int, got %T", result)
			}
			if v.Cmp(tt.want) != 0 {
				t.Fatalf("expected %s, got %s", tt.want.String(), v.String())
			}
		})
	}
}

func TestConvertArg_Uint24(t *testing.T) {
	typ := makeIntType(ethabi.UintTy, 24)
	result, err := ConvertArg("16777215", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := result.(*big.Int)
	if !ok {
		t.Fatalf("expected *big.Int, got %T", result)
	}
	if v.Cmp(big.NewInt(16777215)) != 0 {
		t.Fatalf("expected 16777215, got %s", v.String())
	}
}

func TestConvertArg_Uint128(t *testing.T) {
	typ := makeIntType(ethabi.UintTy, 128)
	result, err := ConvertArg("340282366920938463463374607431768211455", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := result.(*big.Int)
	if !ok {
		t.Fatalf("expected *big.Int, got %T", result)
	}
	maxUint128 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	if v.Cmp(maxUint128) != 0 {
		t.Fatalf("expected max uint128, got %s", v.String())
	}
}

// --- Int tests ---

func TestConvertArg_Int8(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int8
		wantErr bool
	}{
		{"min", "-128", -128, false},
		{"zero", "0", 0, false},
		{"max", "127", 127, false},
		{"overflow_pos", "128", 0, true},
		{"overflow_neg", "-129", 0, true},
		{"invalid", "abc", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := makeIntType(ethabi.IntTy, 8)
			result, err := ConvertArg(tt.input, typ)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result=%v)", result)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			v, ok := result.(int8)
			if !ok {
				t.Fatalf("expected int8, got %T", result)
			}
			if v != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, v)
			}
		})
	}
}

func TestConvertArg_Int16(t *testing.T) {
	typ := makeIntType(ethabi.IntTy, 16)
	result, err := ConvertArg("-32768", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := result.(int16)
	if !ok {
		t.Fatalf("expected int16, got %T", result)
	}
	if v != -32768 {
		t.Fatalf("expected -32768, got %d", v)
	}
}

func TestConvertArg_Int32(t *testing.T) {
	typ := makeIntType(ethabi.IntTy, 32)
	result, err := ConvertArg("2147483647", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := result.(int32)
	if !ok {
		t.Fatalf("expected int32, got %T", result)
	}
	if v != 2147483647 {
		t.Fatalf("expected 2147483647, got %d", v)
	}
}

func TestConvertArg_Int64(t *testing.T) {
	typ := makeIntType(ethabi.IntTy, 64)
	result, err := ConvertArg("-9223372036854775808", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := result.(int64)
	if !ok {
		t.Fatalf("expected int64, got %T", result)
	}
	if v != -9223372036854775808 {
		t.Fatalf("expected min int64, got %d", v)
	}
}

func TestConvertArg_Int256(t *testing.T) {
	minInt256 := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 255))
	maxInt256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(1))

	tests := []struct {
		want    *big.Int
		name    string
		input   string
		wantErr bool
	}{
		{big.NewInt(-1), "negative", "-1", false},
		{minInt256, "min", minInt256.String(), false},
		{maxInt256, "max", maxInt256.String(), false},
		{nil, "overflow_pos", new(big.Int).Add(maxInt256, big.NewInt(1)).String(), true},
		{nil, "overflow_neg", new(big.Int).Sub(minInt256, big.NewInt(1)).String(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := makeIntType(ethabi.IntTy, 256)
			result, err := ConvertArg(tt.input, typ)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result=%v)", result)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			v, ok := result.(*big.Int)
			if !ok {
				t.Fatalf("expected *big.Int, got %T", result)
			}
			if v.Cmp(tt.want) != 0 {
				t.Fatalf("expected %s, got %s", tt.want.String(), v.String())
			}
		})
	}
}

func TestConvertArg_Int40(t *testing.T) {
	typ := makeIntType(ethabi.IntTy, 40)
	result, err := ConvertArg("-549755813888", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := result.(*big.Int)
	if !ok {
		t.Fatalf("expected *big.Int, got %T", result)
	}
	want := big.NewInt(-549755813888)
	if v.Cmp(want) != 0 {
		t.Fatalf("expected %s, got %s", want.String(), v.String())
	}
}

func TestConvertArg_UintHexInput(t *testing.T) {
	typ := makeIntType(ethabi.UintTy, 8)
	result, err := ConvertArg("0x0a", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := result.(uint8)
	if !ok {
		t.Fatalf("expected uint8, got %T", result)
	}
	if v != 10 {
		t.Fatalf("expected 10, got %d", v)
	}
}

func TestConvertArg_UintErrorMessage(t *testing.T) {
	typ := makeIntType(ethabi.UintTy, 8)
	_, err := ConvertArg("abc", typ)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid uint8") {
		t.Fatalf("expected error mentioning 'invalid uint8', got %q", err.Error())
	}
}

func TestConvertArg_IntErrorMessage(t *testing.T) {
	typ := makeIntType(ethabi.IntTy, 256)
	_, err := ConvertArg("not_a_number", typ)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid int256") {
		t.Fatalf("expected error mentioning 'invalid int256', got %q", err.Error())
	}
}
