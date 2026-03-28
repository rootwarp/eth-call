package converter

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
)

// trimHexPrefix removes an optional 0x or 0X prefix from a hex string.
func trimHexPrefix(s string) string {
	s = strings.TrimPrefix(s, "0x")
	return strings.TrimPrefix(s, "0X")
}

// convertBytes converts a hex string to dynamic []byte.
// Accepts optional "0x" prefix.
func convertBytes(s string) ([]byte, error) {
	s = trimHexPrefix(s)

	if s == "" {
		return []byte{}, nil
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid bytes: %q (expected 0x-prefixed hex string): %w", s, err)
	}
	return b, nil
}

// convertFixedBytes converts a hex string to a fixed-size [N]byte array.
// Uses reflect.ArrayOf to construct the exact type required by abi.Pack().
func convertFixedBytes(s string, size int) (interface{}, error) {
	s = trimHexPrefix(s)

	if s == "" {
		arr := reflect.New(reflect.ArrayOf(size, reflect.TypeOf(byte(0)))).Elem()
		return arr.Interface(), nil
	}

	decoded, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid bytes%d: %q (expected 0x-prefixed hex string): %w", size, s, err)
	}

	if len(decoded) > size {
		return nil, fmt.Errorf("invalid bytes%d: input is %d bytes, maximum is %d", size, len(decoded), size)
	}

	arr := reflect.New(reflect.ArrayOf(size, reflect.TypeOf(byte(0)))).Elem()
	for i, b := range decoded {
		arr.Index(i).SetUint(uint64(b))
	}

	return arr.Interface(), nil
}
