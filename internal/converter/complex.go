package converter

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

// convertSlice converts a JSON array string to a Go slice matching the element type.
// Uses json.Decoder.UseNumber() to prevent float64 precision loss.
func convertSlice(s string, elemType ethabi.Type) (interface{}, error) {
	elements, err := parseJSONArray(s)
	if err != nil {
		return nil, fmt.Errorf("invalid array: %w", err)
	}

	goType := elemType.GetType()
	slice := reflect.MakeSlice(reflect.SliceOf(goType), len(elements), len(elements))

	for i, raw := range elements {
		str := elementToString(raw)
		converted, err := ConvertArg(str, elemType)
		if err != nil {
			return nil, fmt.Errorf("element [%d]: %w", i, err)
		}
		slice.Index(i).Set(reflect.ValueOf(converted))
	}

	return slice.Interface(), nil
}

// convertArray converts a JSON array string to a fixed-size Go array matching the element type.
// Validates that the number of elements matches the expected size.
func convertArray(s string, elemType ethabi.Type, size int) (interface{}, error) {
	elements, err := parseJSONArray(s)
	if err != nil {
		return nil, fmt.Errorf("invalid array: %w", err)
	}

	if len(elements) != size {
		return nil, fmt.Errorf("array length mismatch: got %d elements, expected %d", len(elements), size)
	}

	goType := elemType.GetType()
	arr := reflect.New(reflect.ArrayOf(size, goType)).Elem()

	for i, raw := range elements {
		str := elementToString(raw)
		converted, err := ConvertArg(str, elemType)
		if err != nil {
			return nil, fmt.Errorf("element [%d]: %w", i, err)
		}
		arr.Index(i).Set(reflect.ValueOf(converted))
	}

	return arr.Interface(), nil
}

// parseJSONArray decodes a JSON array string into a slice of raw JSON elements.
// Uses json.Decoder.UseNumber() to preserve numeric precision.
func parseJSONArray(s string) ([]json.RawMessage, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()

	var elements []json.RawMessage
	if err := dec.Decode(&elements); err != nil {
		return nil, err
	}
	return elements, nil
}

// convertTuple converts a JSON object string to a dynamic Go struct matching
// the ABI tuple type. Uses json.Decoder.UseNumber() to prevent float64 precision loss.
// Field names are matched using go-ethereum's ToCamelCase convention.
func convertTuple(s string, typ ethabi.Type) (interface{}, error) {
	// Parse JSON object into map of raw messages.
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()

	var fields map[string]json.RawMessage
	if err := dec.Decode(&fields); err != nil {
		return nil, fmt.Errorf("invalid tuple: %w", err)
	}

	// Create a new struct instance using the ABI-defined type.
	structValue := reflect.New(typ.TupleType).Elem()

	for i, elem := range typ.TupleElems {
		rawName := typ.TupleRawNames[i]
		fieldName := ethabi.ToCamelCase(rawName)

		raw, ok := fields[rawName]
		if !ok {
			return nil, fmt.Errorf("tuple field %q missing (expected fields: %s)",
				rawName, strings.Join(typ.TupleRawNames, ", "))
		}

		// Convert the raw JSON value to a string for ConvertArg.
		str := elementToString(raw)
		converted, err := ConvertArg(str, *elem)
		if err != nil {
			return nil, fmt.Errorf("tuple field %q: %w", rawName, err)
		}

		field := structValue.FieldByName(fieldName)
		if !field.IsValid() {
			return nil, fmt.Errorf("tuple field %q: struct field %q not found", rawName, fieldName)
		}
		field.Set(reflect.ValueOf(converted))
	}

	return structValue.Interface(), nil
}

// elementToString converts a raw JSON element to its string representation
// suitable for passing to ConvertArg.
func elementToString(raw json.RawMessage) string {
	// Try to unmarshal as json.Number (for numeric values).
	var num json.Number
	if err := json.Unmarshal(raw, &num); err == nil {
		return num.String()
	}

	// Try to unmarshal as a string (for quoted values like addresses).
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}

	// Try to unmarshal as a bool.
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		if b {
			return "true"
		}
		return "false"
	}

	// For complex types (arrays, objects), return the raw JSON string.
	return string(raw)
}
