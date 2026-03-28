// Package rpc provides an interface for fetching transaction parameters from an Ethereum JSON-RPC endpoint.
package rpc

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/rootwarp/eth-call/internal/txbuilder"
)

// Client defines the interface for fetching transaction parameters from an RPC node.
type Client interface {
	// FetchParams fetches nonce, gas estimates, and chain ID for a transaction.
	FetchParams(ctx context.Context, from common.Address, to common.Address, calldata []byte, value *big.Int) (txbuilder.TxParams, error)

	// Close releases the underlying RPC connection.
	Close()
}

// Dial connects to an Ethereum JSON-RPC endpoint and returns a Client.
func Dial(ctx context.Context, url string) (Client, error) {
	client, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("rpc: failed to connect: %w", err)
	}

	return &ethClient{backend: &realBackend{client: client}, warnWriter: os.Stderr}, nil
}
