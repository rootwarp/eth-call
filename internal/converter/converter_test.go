package converter

import (
	"math/big"
	"reflect"
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

// --- Tuple tests ---

// makeTupleType builds a tuple ethabi.Type using NewType.
func makeTupleType(t *testing.T, components []ethabi.ArgumentMarshaling) ethabi.Type {
	t.Helper()
	typ, err := ethabi.NewType("tuple", "", components)
	if err != nil {
		t.Fatalf("failed to create tuple type: %v", err)
	}
	return typ
}

func TestConvertArg_SimpleTuple(t *testing.T) {
	typ := makeTupleType(t, []ethabi.ArgumentMarshaling{
		{Name: "amount", Type: "uint256"},
		{Name: "recipient", Type: "address"},
	})

	result, err := ConvertArg(`{"amount":"100","recipient":"0xdAC17F958D2ee523a2206206994597C13D831ec7"}`, typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's a struct with correct fields via reflection.
	v := reflect.ValueOf(result)
	if v.Kind() != reflect.Struct {
		t.Fatalf("expected struct, got %s", v.Kind())
	}

	amountField := v.FieldByName("Amount")
	if !amountField.IsValid() {
		t.Fatal("expected field Amount in struct")
	}
	amount, ok := amountField.Interface().(*big.Int)
	if !ok {
		t.Fatalf("expected *big.Int for Amount, got %T", amountField.Interface())
	}
	if amount.Cmp(big.NewInt(100)) != 0 {
		t.Fatalf("expected Amount=100, got %s", amount.String())
	}

	recipientField := v.FieldByName("Recipient")
	if !recipientField.IsValid() {
		t.Fatal("expected field Recipient in struct")
	}
	addr, ok := recipientField.Interface().(common.Address)
	if !ok {
		t.Fatalf("expected common.Address for Recipient, got %T", recipientField.Interface())
	}
	want := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
	if addr != want {
		t.Fatalf("expected Recipient=%s, got %s", want.Hex(), addr.Hex())
	}
}

func TestConvertArg_TupleMissingField(t *testing.T) {
	typ := makeTupleType(t, []ethabi.ArgumentMarshaling{
		{Name: "amount", Type: "uint256"},
		{Name: "recipient", Type: "address"},
	})

	_, err := ConvertArg(`{"amount":"100"}`, typ)
	if err == nil {
		t.Fatal("expected error for missing field, got nil")
	}
	if !strings.Contains(err.Error(), "recipient") && !strings.Contains(err.Error(), "missing") {
		t.Fatalf("expected error mentioning missing field, got %q", err.Error())
	}
}

func TestConvertArg_NestedTuple(t *testing.T) {
	typ := makeTupleType(t, []ethabi.ArgumentMarshaling{
		{Name: "value", Type: "uint256"},
		{Name: "inner", Type: "tuple", Components: []ethabi.ArgumentMarshaling{
			{Name: "addr", Type: "address"},
			{Name: "flag", Type: "bool"},
		}},
	})

	input := `{"value":"42","inner":{"addr":"0xdAC17F958D2ee523a2206206994597C13D831ec7","flag":"true"}}`
	result, err := ConvertArg(input, typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v := reflect.ValueOf(result)
	valueField := v.FieldByName("Value")
	if !valueField.IsValid() {
		t.Fatal("expected field Value")
	}
	val, ok := valueField.Interface().(*big.Int)
	if !ok {
		t.Fatalf("expected *big.Int, got %T", valueField.Interface())
	}
	if val.Cmp(big.NewInt(42)) != 0 {
		t.Fatalf("expected Value=42, got %s", val.String())
	}

	innerField := v.FieldByName("Inner")
	if !innerField.IsValid() {
		t.Fatal("expected field Inner")
	}
	if innerField.Kind() != reflect.Struct {
		t.Fatalf("expected Inner to be struct, got %s", innerField.Kind())
	}

	addrField := innerField.FieldByName("Addr")
	if !addrField.IsValid() {
		t.Fatal("expected field Addr in Inner")
	}
	addr, ok := addrField.Interface().(common.Address)
	if !ok {
		t.Fatalf("expected common.Address, got %T", addrField.Interface())
	}
	wantAddr := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
	if addr != wantAddr {
		t.Fatalf("expected addr=%s, got %s", wantAddr.Hex(), addr.Hex())
	}

	flagField := innerField.FieldByName("Flag")
	if !flagField.IsValid() {
		t.Fatal("expected field Flag in Inner")
	}
	if flagField.Bool() != true {
		t.Fatal("expected Flag=true")
	}
}

func TestConvertArg_ArrayOfTuples(t *testing.T) {
	tupleType := makeTupleType(t, []ethabi.ArgumentMarshaling{
		{Name: "to", Type: "address"},
		{Name: "amount", Type: "uint256"},
	})
	sliceType := makeSliceType(tupleType)

	input := `[{"to":"0xdAC17F958D2ee523a2206206994597C13D831ec7","amount":"100"},{"to":"0x0000000000000000000000000000000000000001","amount":"200"}]`
	result, err := ConvertArg(input, sliceType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v := reflect.ValueOf(result)
	if v.Kind() != reflect.Slice {
		t.Fatalf("expected slice, got %s", v.Kind())
	}
	if v.Len() != 2 {
		t.Fatalf("expected 2 elements, got %d", v.Len())
	}

	elem0 := v.Index(0)
	amount0, ok := elem0.FieldByName("Amount").Interface().(*big.Int)
	if !ok {
		t.Fatalf("expected *big.Int, got %T", elem0.FieldByName("Amount").Interface())
	}
	if amount0.Cmp(big.NewInt(100)) != 0 {
		t.Fatalf("expected element[0].Amount=100, got %s", amount0.String())
	}

	elem1 := v.Index(1)
	amount1, ok := elem1.FieldByName("Amount").Interface().(*big.Int)
	if !ok {
		t.Fatalf("expected *big.Int, got %T", elem1.FieldByName("Amount").Interface())
	}
	if amount1.Cmp(big.NewInt(200)) != 0 {
		t.Fatalf("expected element[1].Amount=200, got %s", amount1.String())
	}
}

func TestConvertArg_TupleInvalidJSON(t *testing.T) {
	typ := makeTupleType(t, []ethabi.ArgumentMarshaling{
		{Name: "value", Type: "uint256"},
	})

	_, err := ConvertArg("not json", typ)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestConvertArg_TupleFieldConversionError(t *testing.T) {
	typ := makeTupleType(t, []ethabi.ArgumentMarshaling{
		{Name: "value", Type: "uint8"},
	})

	_, err := ConvertArg(`{"value":"999"}`, typ)
	if err == nil {
		t.Fatal("expected error for field conversion failure, got nil")
	}
	if !strings.Contains(err.Error(), "value") {
		t.Fatalf("expected error mentioning field name, got %q", err.Error())
	}
}

// --- Slice (dynamic array) tests ---

// makeSliceType builds an ethabi.Type for T[] where T is described by elemType.
func makeSliceType(elemType ethabi.Type) ethabi.Type {
	return ethabi.Type{T: ethabi.SliceTy, Elem: &elemType}
}

// makeArrayType builds an ethabi.Type for T[N].
func makeArrayType(elemType ethabi.Type, size int) ethabi.Type {
	return ethabi.Type{T: ethabi.ArrayTy, Elem: &elemType, Size: size}
}

func TestConvertArg_Uint256Slice(t *testing.T) {
	elem := makeIntType(ethabi.UintTy, 256)
	typ := makeSliceType(elem)

	t.Run("basic", func(t *testing.T) {
		result, err := ConvertArg("[1,2,3]", typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		slice, ok := result.([]*big.Int)
		if !ok {
			t.Fatalf("expected []*big.Int, got %T", result)
		}
		if len(slice) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(slice))
		}
		for i, want := range []int64{1, 2, 3} {
			if slice[i].Cmp(big.NewInt(want)) != 0 {
				t.Fatalf("element %d: expected %d, got %s", i, want, slice[i].String())
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		result, err := ConvertArg("[]", typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		slice, ok := result.([]*big.Int)
		if !ok {
			t.Fatalf("expected []*big.Int, got %T", result)
		}
		if len(slice) != 0 {
			t.Fatalf("expected 0 elements, got %d", len(slice))
		}
	})

	t.Run("single", func(t *testing.T) {
		result, err := ConvertArg("[42]", typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		slice, ok := result.([]*big.Int)
		if !ok {
			t.Fatalf("expected []*big.Int, got %T", result)
		}
		if len(slice) != 1 || slice[0].Cmp(big.NewInt(42)) != 0 {
			t.Fatalf("expected [42], got %v", slice)
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		_, err := ConvertArg("not json", typ)
		if err == nil {
			t.Fatal("expected error for invalid JSON, got nil")
		}
	})
}

func TestConvertArg_AddressSlice(t *testing.T) {
	elem := makeType(ethabi.AddressTy)
	typ := makeSliceType(elem)

	result, err := ConvertArg(`["0xdAC17F958D2ee523a2206206994597C13D831ec7","0x0000000000000000000000000000000000000001"]`, typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	slice, ok := result.([]common.Address)
	if !ok {
		t.Fatalf("expected []common.Address, got %T", result)
	}
	if len(slice) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(slice))
	}
	want0 := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
	if slice[0] != want0 {
		t.Fatalf("element 0: expected %s, got %s", want0.Hex(), slice[0].Hex())
	}
}

func TestConvertArg_BoolSlice(t *testing.T) {
	elem := makeType(ethabi.BoolTy)
	typ := makeSliceType(elem)

	result, err := ConvertArg("[true,false,true]", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	slice, ok := result.([]bool)
	if !ok {
		t.Fatalf("expected []bool, got %T", result)
	}
	if len(slice) != 3 || slice[0] != true || slice[1] != false || slice[2] != true {
		t.Fatalf("expected [true,false,true], got %v", slice)
	}
}

func TestConvertArg_StringSlice(t *testing.T) {
	elem := makeType(ethabi.StringTy)
	typ := makeSliceType(elem)

	result, err := ConvertArg(`["hello","world"]`, typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	slice, ok := result.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", result)
	}
	if len(slice) != 2 || slice[0] != "hello" || slice[1] != "world" {
		t.Fatalf("expected [hello,world], got %v", slice)
	}
}

func TestConvertArg_SliceElementError(t *testing.T) {
	elem := makeIntType(ethabi.UintTy, 8)
	typ := makeSliceType(elem)

	_, err := ConvertArg(`[1,256,3]`, typ)
	if err == nil {
		t.Fatal("expected error for element overflow, got nil")
	}
	if !strings.Contains(err.Error(), "element [1]") {
		t.Fatalf("expected error identifying element index, got %q", err.Error())
	}
}

// --- Fixed array tests ---

func TestConvertArg_Uint256Array2(t *testing.T) {
	elem := makeIntType(ethabi.UintTy, 256)
	typ := makeArrayType(elem, 2)

	t.Run("valid", func(t *testing.T) {
		result, err := ConvertArg("[1,2]", typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr, ok := result.([2]*big.Int)
		if !ok {
			t.Fatalf("expected [2]*big.Int, got %T", result)
		}
		if arr[0].Cmp(big.NewInt(1)) != 0 || arr[1].Cmp(big.NewInt(2)) != 0 {
			t.Fatalf("expected [1,2], got [%s,%s]", arr[0], arr[1])
		}
	})

	t.Run("length_mismatch", func(t *testing.T) {
		_, err := ConvertArg("[1,2,3]", typ)
		if err == nil {
			t.Fatal("expected error for length mismatch, got nil")
		}
		if !strings.Contains(err.Error(), "length mismatch") {
			t.Fatalf("expected length mismatch error, got %q", err.Error())
		}
	})
}

func TestConvertArg_FixedArrayInvalidJSON(t *testing.T) {
	elem := makeIntType(ethabi.UintTy, 256)
	typ := makeArrayType(elem, 2)

	_, err := ConvertArg("not json", typ)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestConvertArg_FixedArrayElementError(t *testing.T) {
	elem := makeIntType(ethabi.UintTy, 8)
	typ := makeArrayType(elem, 2)

	_, err := ConvertArg("[1,999]", typ)
	if err == nil {
		t.Fatal("expected error for element overflow, got nil")
	}
	if !strings.Contains(err.Error(), "element [1]") {
		t.Fatalf("expected error identifying element index, got %q", err.Error())
	}
}

func TestConvertArg_Uint8Array(t *testing.T) {
	elem := makeIntType(ethabi.UintTy, 8)
	typ := makeArrayType(elem, 3)

	result, err := ConvertArg("[10,20,30]", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	arr, ok := result.([3]uint8)
	if !ok {
		t.Fatalf("expected [3]uint8, got %T", result)
	}
	if arr[0] != 10 || arr[1] != 20 || arr[2] != 30 {
		t.Fatalf("expected [10,20,30], got %v", arr)
	}
}

func TestConvertArg_Uint256LargeNumbersSlice(t *testing.T) {
	elem := makeIntType(ethabi.UintTy, 256)
	typ := makeSliceType(elem)

	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	input := "[" + maxUint256.String() + "]"
	result, err := ConvertArg(input, typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	slice, ok := result.([]*big.Int)
	if !ok {
		t.Fatalf("expected []*big.Int, got %T", result)
	}
	if len(slice) != 1 || slice[0].Cmp(maxUint256) != 0 {
		t.Fatalf("expected [%s], got %v", maxUint256.String(), slice)
	}
}

// --- Dynamic bytes tests ---

func TestConvertArg_Bytes(t *testing.T) {
	typ := makeType(ethabi.BytesTy)

	t.Run("valid_hex", func(t *testing.T) {
		result, err := ConvertArg("0xdeadbeef", typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, ok := result.([]byte)
		if !ok {
			t.Fatalf("expected []byte, got %T", result)
		}
		if len(b) != 4 {
			t.Fatalf("expected 4 bytes, got %d", len(b))
		}
		if b[0] != 0xde || b[1] != 0xad || b[2] != 0xbe || b[3] != 0xef {
			t.Fatalf("expected deadbeef, got %x", b)
		}
	})

	t.Run("empty_0x", func(t *testing.T) {
		result, err := ConvertArg("0x", typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, ok := result.([]byte)
		if !ok {
			t.Fatalf("expected []byte, got %T", result)
		}
		if len(b) != 0 {
			t.Fatalf("expected empty bytes, got %d bytes", len(b))
		}
	})

	t.Run("no_prefix", func(t *testing.T) {
		result, err := ConvertArg("abcd", typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, ok := result.([]byte)
		if !ok {
			t.Fatalf("expected []byte, got %T", result)
		}
		if len(b) != 2 || b[0] != 0xab || b[1] != 0xcd {
			t.Fatalf("expected abcd, got %x", b)
		}
	})

	t.Run("odd_length", func(t *testing.T) {
		_, err := ConvertArg("0xabc", typ)
		if err == nil {
			t.Fatal("expected error for odd-length hex, got nil")
		}
		if !strings.Contains(err.Error(), "invalid bytes") {
			t.Fatalf("expected 'invalid bytes' error, got %q", err.Error())
		}
	})

	t.Run("invalid_hex", func(t *testing.T) {
		_, err := ConvertArg("0xZZZZ", typ)
		if err == nil {
			t.Fatal("expected error for invalid hex, got nil")
		}
		if !strings.Contains(err.Error(), "invalid bytes") {
			t.Fatalf("expected 'invalid bytes' error, got %q", err.Error())
		}
	})

	t.Run("empty_string", func(t *testing.T) {
		result, err := ConvertArg("", typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, ok := result.([]byte)
		if !ok {
			t.Fatalf("expected []byte, got %T", result)
		}
		if len(b) != 0 {
			t.Fatalf("expected empty bytes, got %d bytes", len(b))
		}
	})
}

// --- Fixed bytes tests ---

func TestConvertArg_Bytes1(t *testing.T) {
	typ := makeIntType(ethabi.FixedBytesTy, 1)

	t.Run("valid", func(t *testing.T) {
		result, err := ConvertArg("0xab", typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr, ok := result.([1]byte)
		if !ok {
			t.Fatalf("expected [1]byte, got %T", result)
		}
		if arr[0] != 0xab {
			t.Fatalf("expected 0xab, got 0x%x", arr[0])
		}
	})

	t.Run("too_long", func(t *testing.T) {
		_, err := ConvertArg("0xabcd", typ)
		if err == nil {
			t.Fatal("expected error for too-long input, got nil")
		}
		if !strings.Contains(err.Error(), "invalid bytes1") {
			t.Fatalf("expected 'invalid bytes1' error, got %q", err.Error())
		}
	})
}

func TestConvertArg_Bytes4(t *testing.T) {
	typ := makeIntType(ethabi.FixedBytesTy, 4)

	t.Run("function_selector", func(t *testing.T) {
		result, err := ConvertArg("0xa9059cbb", typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr, ok := result.([4]byte)
		if !ok {
			t.Fatalf("expected [4]byte, got %T", result)
		}
		expected := [4]byte{0xa9, 0x05, 0x9c, 0xbb}
		if arr != expected {
			t.Fatalf("expected %x, got %x", expected, arr)
		}
	})
}

func TestConvertArg_Bytes32(t *testing.T) {
	typ := makeIntType(ethabi.FixedBytesTy, 32)

	t.Run("full_32_bytes", func(t *testing.T) {
		input := "0x" + strings.Repeat("ab", 32)
		result, err := ConvertArg(input, typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr, ok := result.([32]byte)
		if !ok {
			t.Fatalf("expected [32]byte, got %T", result)
		}
		for i := 0; i < 32; i++ {
			if arr[i] != 0xab {
				t.Fatalf("byte %d: expected 0xab, got 0x%x", i, arr[i])
			}
		}
	})

	t.Run("shorter_input_left_padded", func(t *testing.T) {
		result, err := ConvertArg("0xab", typ)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr, ok := result.([32]byte)
		if !ok {
			t.Fatalf("expected [32]byte, got %T", result)
		}
		if arr[0] != 0xab {
			t.Fatalf("expected first byte 0xab, got 0x%x", arr[0])
		}
		for i := 1; i < 32; i++ {
			if arr[i] != 0 {
				t.Fatalf("byte %d: expected 0x00, got 0x%x", i, arr[i])
			}
		}
	})

	t.Run("too_long_33_bytes", func(t *testing.T) {
		input := "0x" + strings.Repeat("ab", 33)
		_, err := ConvertArg(input, typ)
		if err == nil {
			t.Fatal("expected error for 33 bytes in bytes32, got nil")
		}
		if !strings.Contains(err.Error(), "invalid bytes32") {
			t.Fatalf("expected 'invalid bytes32' error, got %q", err.Error())
		}
	})

	t.Run("invalid_hex", func(t *testing.T) {
		_, err := ConvertArg("0xZZZZ", typ)
		if err == nil {
			t.Fatal("expected error for invalid hex, got nil")
		}
		if !strings.Contains(err.Error(), "invalid bytes32") {
			t.Fatalf("expected 'invalid bytes32' error, got %q", err.Error())
		}
	})
}

func TestConvertArg_Bytes20(t *testing.T) {
	typ := makeIntType(ethabi.FixedBytesTy, 20)

	result, err := ConvertArg("0x"+strings.Repeat("ff", 20), typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	arr, ok := result.([20]byte)
	if !ok {
		t.Fatalf("expected [20]byte, got %T", result)
	}
	for i := 0; i < 20; i++ {
		if arr[i] != 0xff {
			t.Fatalf("byte %d: expected 0xff, got 0x%x", i, arr[i])
		}
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
