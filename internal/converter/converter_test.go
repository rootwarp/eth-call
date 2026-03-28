package converter

import (
	"strings"
	"testing"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

// makeType is a test helper that creates an ethabi.Type with the given T value.
func makeType(t byte) ethabi.Type {
	return ethabi.Type{T: t}
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
		{"UintTy", ethabi.UintTy},
		{"IntTy", ethabi.IntTy},
		{"BoolTy", ethabi.BoolTy},
		{"AddressTy", ethabi.AddressTy},
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
