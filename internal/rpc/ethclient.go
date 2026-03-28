package rpc

import (
	"context"
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum"
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
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
	backend    ethClientBackend
	warnWriter io.Writer
}

// FetchParams fetches transaction parameters from the RPC node.
// Individual RPC call failures degrade gracefully with warnings, except
// PendingNonceAt which always returns a hard error (nonce must be correct).
func (c *ethClient) FetchParams(ctx context.Context, from, to common.Address, calldata []byte, value *big.Int) (txbuilder.TxParams, error) {
	// Chain ID: degradable — nil signals caller to use its own default
	chainID, err := c.backend.ChainID(ctx)
	if err != nil {
		c.warn("chain ID", err)
		chainID = nil
	}

	// Nonce: NOT degradable — must be correct
	nonce, err := c.backend.PendingNonceAt(ctx, from)
	if err != nil {
		return txbuilder.TxParams{}, fmt.Errorf("rpc: failed to fetch nonce: %w", err)
	}

	// Gas tip cap: degradable — default to 0
	gasTipCap, err := c.backend.SuggestGasTipCap(ctx)
	if err != nil {
		c.warn("gas tip cap", err)
		gasTipCap = big.NewInt(0)
	}

	// Base fee (via header): degradable — default gas fee cap to 0
	var gasFeeCap *big.Int
	header, err := c.backend.HeaderByNumber(ctx, nil)
	if err != nil {
		c.warn("base fee", err)
		gasFeeCap = big.NewInt(0)
	} else {
		gasFeeCap = computeGasFeeCap(header.BaseFee, gasTipCap)
	}

	// Gas estimate: degradable — default to 0, extract revert reason if available
	gasLimit, err := c.backend.EstimateGas(ctx, ethereum.CallMsg{
		From:      from,
		To:        &to,
		Data:      calldata,
		Value:     value,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
	})
	if err != nil {
		c.warnEstimateGas(err)
		gasLimit = 0
	} else {
		gasLimit = applyGasMargin(gasLimit)
	}

	return txbuilder.TxParams{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		GasLimit:  gasLimit,
		Value:     value,
	}, nil
}

// Close releases the underlying RPC connection.
func (c *ethClient) Close() {
	c.backend.Close()
}

// warn emits a warning about a failed RPC call to the warn writer.
func (c *ethClient) warn(field string, err error) {
	if c.warnWriter != nil {
		_, _ = fmt.Fprintf(c.warnWriter, "warning: failed to fetch %s, using default: %v\n", field, err)
	}
}

// warnEstimateGas emits a warning about a failed gas estimation,
// extracting the revert reason if available.
func (c *ethClient) warnEstimateGas(err error) {
	if c.warnWriter == nil {
		return
	}

	if reason := extractRevertReason(err); reason != "" {
		_, _ = fmt.Fprintf(c.warnWriter, "warning: failed to estimate gas, contract reverted: %s\n", reason)
		return
	}

	_, _ = fmt.Fprintf(c.warnWriter, "warning: failed to estimate gas, using default: %v\n", err)
}

// extractRevertReason tries to extract a revert reason from an error.
// Returns the reason string or empty string if extraction fails.
func extractRevertReason(err error) string {
	data, ok := ethclient.RevertErrorData(err)
	if !ok {
		return ""
	}

	reason, unpackErr := ethabi.UnpackRevert(data)
	if unpackErr != nil {
		return ""
	}

	return reason
}

// computeGasFeeCap calculates the gas fee cap as 2*baseFee + gasTipCap.
func computeGasFeeCap(baseFee, gasTipCap *big.Int) *big.Int {
	feeCap := new(big.Int).Mul(baseFee, big.NewInt(2))
	feeCap.Add(feeCap, gasTipCap)
	return feeCap
}

// applyGasMargin adds a 20% safety margin to the gas limit.
func applyGasMargin(gasLimit uint64) uint64 {
	return gasLimit * 120 / 100
}
