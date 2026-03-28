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

// applyDefaults fills nil fields with default values.
// nil ChainID → 1, nil GasTipCap/GasFeeCap/Value → 0.
func applyDefaults(params *TxParams) {
	if params.ChainID == nil {
		params.ChainID = big.NewInt(1)
	}
	if params.GasTipCap == nil {
		params.GasTipCap = big.NewInt(0)
	}
	if params.GasFeeCap == nil {
		params.GasFeeCap = big.NewInt(0)
	}
	if params.Value == nil {
		params.Value = big.NewInt(0)
	}
}

// buildDynamicFeeTx constructs the DynamicFeeTx from params, calldata, and to address.
func buildDynamicFeeTx(calldata []byte, to common.Address, params TxParams) *types.Transaction {
	applyDefaults(&params)

	txData := &types.DynamicFeeTx{
		ChainID:    params.ChainID,
		Nonce:      params.Nonce,
		GasTipCap:  params.GasTipCap,
		GasFeeCap:  params.GasFeeCap,
		Gas:        params.GasLimit,
		To:         &to,
		Value:      params.Value,
		Data:       calldata,
		AccessList: types.AccessList{},
	}

	return types.NewTx(txData)
}

// Build constructs an unsigned EIP-1559 transaction and returns it
// as an RLP-serialized, 0x-prefixed hex string.
func Build(calldata []byte, to common.Address, params TxParams) (string, error) {
	tx := buildDynamicFeeTx(calldata, to, params)

	raw, err := tx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("txbuilder: serialize failed: %w", err)
	}

	return fmt.Sprintf("0x%x", raw), nil
}

// BuildTx constructs and returns the unsigned *types.Transaction without serializing.
func BuildTx(calldata []byte, to common.Address, params TxParams) (*types.Transaction, error) {
	return buildDynamicFeeTx(calldata, to, params), nil
}
