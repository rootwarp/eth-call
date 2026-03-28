package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rootwarp/eth-call/internal/txbuilder"
)

// MockClient is a configurable mock implementation of Client for testing.
// Set the function fields to control behavior per test.
type MockClient struct {
	FetchParamsFn func(ctx context.Context, from, to common.Address, calldata []byte, value *big.Int) (txbuilder.TxParams, error)
	CloseFn       func()
}

// FetchParams calls FetchParamsFn if set, otherwise returns zero TxParams.
func (m *MockClient) FetchParams(ctx context.Context, from, to common.Address, calldata []byte, value *big.Int) (txbuilder.TxParams, error) {
	if m.FetchParamsFn != nil {
		return m.FetchParamsFn(ctx, from, to, calldata, value)
	}
	return txbuilder.TxParams{}, nil
}

// Close calls CloseFn if set, otherwise does nothing.
func (m *MockClient) Close() {
	if m.CloseFn != nil {
		m.CloseFn()
	}
}
