// Package encoder encodes method calls into ABI-packed calldata bytes.
package encoder

import (
	"fmt"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

// Encode packs a method call with the given arguments into calldata.
func Encode(parsedABI ethabi.ABI, method string, args []interface{}) ([]byte, error) {
	_ = parsedABI
	_ = method
	_ = args
	return nil, fmt.Errorf("not implemented")
}

// EncodeToHex is a convenience wrapper that returns "0x"-prefixed hex.
func EncodeToHex(parsedABI ethabi.ABI, method string, args []interface{}) (string, error) {
	_ = parsedABI
	_ = method
	_ = args
	return "", fmt.Errorf("not implemented")
}
