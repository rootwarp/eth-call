// Package encoder encodes method calls into ABI-packed calldata bytes.
package encoder

import (
	"encoding/hex"
	"fmt"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

// Encode packs a method call with the given arguments into calldata.
// The returned bytes contain the 4-byte method selector followed by
// ABI-encoded arguments.
func Encode(parsedABI ethabi.ABI, method string, args []interface{}) ([]byte, error) {
	calldata, err := parsedABI.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("encode: pack failed for method %q with %d args: %w", method, len(args), err)
	}
	return calldata, nil
}

// EncodeToHex is a convenience wrapper that returns "0x"-prefixed hex.
func EncodeToHex(parsedABI ethabi.ABI, method string, args []interface{}) (string, error) {
	calldata, err := Encode(parsedABI, method, args)
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(calldata), nil
}
