package converter

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// convertUint converts a string to the Go type expected by abi.Pack for unsigned integers.
// Returns uint8/16/32/64 for sizes 8/16/32/64, *big.Int for all other sizes.
func convertUint(s string, size int) (interface{}, error) {
	switch size {
	case 8, 16, 32, 64:
		n, err := strconv.ParseUint(s, 0, size)
		if err != nil {
			return nil, fmt.Errorf("invalid uint%d: %q (expected decimal or 0x-prefixed hex integer)", size, s)
		}
		switch size {
		case 8:
			return uint8(n), nil //nolint:gosec // ParseUint with size=8 guarantees n fits in uint8
		case 16:
			return uint16(n), nil //nolint:gosec // ParseUint with size=16 guarantees n fits in uint16
		case 32:
			return uint32(n), nil //nolint:gosec // ParseUint with size=32 guarantees n fits in uint32
		default:
			return n, nil
		}
	default:
		n, ok := parseBigInt(s)
		if !ok {
			return nil, fmt.Errorf("invalid uint%d: %q (expected decimal or 0x-prefixed hex integer)", size, s)
		}
		if n.Sign() < 0 {
			return nil, fmt.Errorf("invalid uint%d: %q (value must not be negative)", size, s)
		}
		if n.BitLen() > size {
			return nil, fmt.Errorf("invalid uint%d: %q (value overflows %d bits)", size, s, size)
		}
		return n, nil
	}
}

// convertInt converts a string to the Go type expected by abi.Pack for signed integers.
// Returns int8/16/32/64 for sizes 8/16/32/64, *big.Int for all other sizes.
func convertInt(s string, size int) (interface{}, error) {
	switch size {
	case 8, 16, 32, 64:
		n, err := strconv.ParseInt(s, 0, size)
		if err != nil {
			return nil, fmt.Errorf("invalid int%d: %q (expected decimal or 0x-prefixed hex integer)", size, s)
		}
		switch size {
		case 8:
			return int8(n), nil //nolint:gosec // ParseInt with size=8 guarantees n fits in int8
		case 16:
			return int16(n), nil //nolint:gosec // ParseInt with size=16 guarantees n fits in int16
		case 32:
			return int32(n), nil //nolint:gosec // ParseInt with size=32 guarantees n fits in int32
		default:
			return n, nil
		}
	default:
		n, ok := parseBigInt(s)
		if !ok {
			return nil, fmt.Errorf("invalid int%d: %q (expected decimal or 0x-prefixed hex integer)", size, s)
		}
		// Check range: -(2^(size-1)) to 2^(size-1)-1
		bits := uint(size - 1) //nolint:gosec // size is always positive from go-ethereum ABI
		maxVal := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), bits), big.NewInt(1))
		minVal := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), bits))
		if n.Cmp(maxVal) > 0 || n.Cmp(minVal) < 0 {
			return nil, fmt.Errorf("invalid int%d: %q (value out of range [%s, %s])", size, s, minVal.String(), maxVal.String())
		}
		return n, nil
	}
}

// parseBigInt parses a string as a big.Int, auto-detecting hex (0x prefix) or decimal.
func parseBigInt(s string) (*big.Int, bool) {
	n := new(big.Int)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		_, ok := n.SetString(s[2:], 16)
		return n, ok
	}
	_, ok := n.SetString(s, 10)
	return n, ok
}
