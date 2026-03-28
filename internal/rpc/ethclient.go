package rpc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/rootwarp/eth-call/internal/txbuilder"
)

// ethClientBackend abstracts the ethclient methods used by ethClient,
// enabling mock-based testing without a real RPC connection.
type ethClientBackend interface {
	ChainID(ctx context.Context) (*big.Int, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	Close()
}

// realBackend wraps *ethclient.Client to satisfy ethClientBackend.
type realBackend struct {
	client *ethclient.Client
}

func (r *realBackend) ChainID(ctx context.Context) (*big.Int, error) {
	return r.client.ChainID(ctx)
}

func (r *realBackend) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return r.client.PendingNonceAt(ctx, account)
}

func (r *realBackend) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return r.client.SuggestGasTipCap(ctx)
}

func (r *realBackend) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return r.client.HeaderByNumber(ctx, number)
}

func (r *realBackend) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return r.client.EstimateGas(ctx, msg)
}

func (r *realBackend) Close() {
	r.client.Close()
}

// ethClient implements Client using an ethClientBackend.
type ethClient struct {
	backend ethClientBackend
}

// FetchParams fetches transaction parameters from the RPC node.
func (c *ethClient) FetchParams(ctx context.Context, from, to common.Address, calldata []byte, value *big.Int) (txbuilder.TxParams, error) {
	chainID, err := c.backend.ChainID(ctx)
	if err != nil {
		return txbuilder.TxParams{}, fmt.Errorf("rpc: failed to fetch chain ID: %w", err)
	}

	nonce, err := c.backend.PendingNonceAt(ctx, from)
	if err != nil {
		return txbuilder.TxParams{}, fmt.Errorf("rpc: failed to fetch nonce: %w", err)
	}

	gasTipCap, err := c.backend.SuggestGasTipCap(ctx)
	if err != nil {
		return txbuilder.TxParams{}, fmt.Errorf("rpc: failed to fetch gas tip cap: %w", err)
	}

	header, err := c.backend.HeaderByNumber(ctx, nil)
	if err != nil {
		return txbuilder.TxParams{}, fmt.Errorf("rpc: failed to fetch base fee: %w", err)
	}

	gasFeeCap := computeGasFeeCap(header.BaseFee, gasTipCap)

	gasLimit, err := c.backend.EstimateGas(ctx, ethereum.CallMsg{
		From:      from,
		To:        &to,
		Data:      calldata,
		Value:     value,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
	})
	if err != nil {
		return txbuilder.TxParams{}, fmt.Errorf("rpc: failed to estimate gas: %w", err)
	}

	return txbuilder.TxParams{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		GasLimit:  applyGasMargin(gasLimit),
		Value:     value,
	}, nil
}

// Close releases the underlying RPC connection.
func (c *ethClient) Close() {
	c.backend.Close()
}

// computeGasFeeCap calculates the gas fee cap as 2*baseFee + gasTipCap.
func computeGasFeeCap(baseFee, gasTipCap *big.Int) *big.Int {
	// gasFeeCap = 2 * baseFee + gasTipCap
	feeCap := new(big.Int).Mul(baseFee, big.NewInt(2))
	feeCap.Add(feeCap, gasTipCap)
	return feeCap
}

// applyGasMargin adds a 20% safety margin to the gas limit.
func applyGasMargin(gasLimit uint64) uint64 {
	return gasLimit * 120 / 100
}
