package converter

import (
	"fmt"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

// ConvertArgs converts a slice of string arguments to Go types matching
// the ABI method's input parameter types.
func ConvertArgs(args []string, inputs ethabi.Arguments) ([]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

// ConvertArg converts a single string argument to the Go type specified
// by the ABI type descriptor.
func ConvertArg(value string, typ ethabi.Type) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
