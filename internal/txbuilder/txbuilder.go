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
	Nonce     uint64
	GasTipCap *big.Int
	GasFeeCap *big.Int
	GasLimit  uint64
	Value     *big.Int
}

// Build constructs an unsigned EIP-1559 transaction and returns it
// as an RLP-serialized, 0x-prefixed hex string.
func Build(calldata []byte, to common.Address, params TxParams) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// BuildTx constructs and returns the unsigned *types.Transaction without serializing.
func BuildTx(calldata []byte, to common.Address, params TxParams) (*types.Transaction, error) {
	return nil, fmt.Errorf("not implemented")
}
