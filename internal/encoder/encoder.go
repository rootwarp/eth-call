package encoder

import (
	"fmt"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

// Encode packs a method call with the given arguments into calldata.
func Encode(parsedABI ethabi.ABI, method string, args []interface{}) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

// EncodeToHex is a convenience wrapper that returns "0x"-prefixed hex.
func EncodeToHex(parsedABI ethabi.ABI, method string, args []interface{}) (string, error) {
	return "", fmt.Errorf("not implemented")
}
