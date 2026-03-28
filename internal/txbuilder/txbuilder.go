// Package txbuilder constructs unsigned EIP-1559 transactions and serializes them to hex.
package txbuilder

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TxParams holds transaction parameters. Fields left as zero/nil use defaults.
type TxParams struct {
	ChainID   *big.Int
	GasTipCap *big.Int
	GasFeeCap *big.Int
	Value     *big.Int
	Nonce     uint64
	GasLimit  uint64
}

// Build constructs an unsigned EIP-1559 transaction and returns it
// as an RLP-serialized, 0x-prefixed hex string.
func Build(calldata []byte, to common.Address, params TxParams) (string, error) {
	_ = calldata
	_ = to
	_ = params
	return "", fmt.Errorf("not implemented")
}

// BuildTx constructs and returns the unsigned *types.Transaction without serializing.
func BuildTx(calldata []byte, to common.Address, params TxParams) (*types.Transaction, error) {
	_ = calldata
	_ = to
	_ = params
	return nil, fmt.Errorf("not implemented")
}
