// Package converter converts CLI string arguments to Go types for ABI encoding.
package converter

import (
	"fmt"
	"strconv"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// ConvertArgs converts a slice of string arguments to Go types matching
// the ABI method's input parameter types. Returns an error identifying
// the specific argument index and expected type on failure.
func ConvertArgs(args []string, inputs ethabi.Arguments) ([]interface{}, error) {
	if len(args) != len(inputs) {
		return nil, fmt.Errorf("converter: expected %d arguments, got %d", len(inputs), len(args))
	}

	result := make([]interface{}, len(args))
	for i, input := range inputs {
		converted, err := ConvertArg(args[i], input.Type)
		if err != nil {
			return nil, fmt.Errorf("converter: arg[%d] (%s): %w", i, input.Name, err)
		}
		result[i] = converted
	}
	return result, nil
}

// ConvertArg converts a single string argument to the Go type specified
// by the ABI type descriptor. Handles recursive types (tuples, arrays of tuples).
func ConvertArg(value string, typ ethabi.Type) (interface{}, error) {
	switch typ.T {
	case ethabi.StringTy:
		return value, nil
	case ethabi.UintTy:
		return convertUint(value, typ.Size)
	case ethabi.IntTy:
		return convertInt(value, typ.Size)
	case ethabi.BoolTy:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("invalid bool: %q (expected true, false, 1, or 0)", value)
		}
		return b, nil
	case ethabi.AddressTy:
		if !common.IsHexAddress(value) {
			return nil, fmt.Errorf("invalid address: %q (expected 0x-prefixed 40-character hex string)", value)
		}
		return common.HexToAddress(value), nil
	case ethabi.BytesTy:
		return convertBytes(value)
	case ethabi.FixedBytesTy:
		return convertFixedBytes(value, typ.Size)
	case ethabi.SliceTy:
		return nil, fmt.Errorf("converter: %s not implemented", typ.String())
	case ethabi.ArrayTy:
		return nil, fmt.Errorf("converter: %s not implemented", typ.String())
	case ethabi.TupleTy:
		return nil, fmt.Errorf("converter: %s not implemented", typ.String())
	default:
		return nil, fmt.Errorf("converter: unsupported type: %s", typ.String())
	}
}
