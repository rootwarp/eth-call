package rpc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// mockBackend implements ethClientBackend for testing.
type mockBackend struct {
	chainID        *big.Int
	chainIDErr     error
	gasTipCap      *big.Int
	gasTipCapErr   error
	baseFee        *big.Int
	headerErr      error
	gasEstimateErr error
	nonceErr       error
	nonce          uint64
	gasEstimate    uint64
}

func (m *mockBackend) ChainID(_ context.Context) (*big.Int, error) {
	if m.chainIDErr != nil {
		return nil, m.chainIDErr
	}
	return m.chainID, nil
}

func (m *mockBackend) PendingNonceAt(_ context.Context, _ common.Address) (uint64, error) {
	if m.nonceErr != nil {
		return 0, m.nonceErr
	}
	return m.nonce, nil
}

func (m *mockBackend) SuggestGasTipCap(_ context.Context) (*big.Int, error) {
	if m.gasTipCapErr != nil {
		return nil, m.gasTipCapErr
	}
	return m.gasTipCap, nil
}

func (m *mockBackend) HeaderByNumber(_ context.Context, _ *big.Int) (*types.Header, error) {
	if m.headerErr != nil {
		return nil, m.headerErr
	}
	return &types.Header{BaseFee: m.baseFee}, nil
}

func (m *mockBackend) EstimateGas(_ context.Context, _ ethereum.CallMsg) (uint64, error) {
	if m.gasEstimateErr != nil {
		return 0, m.gasEstimateErr
	}
	return m.gasEstimate, nil
}

func (m *mockBackend) Close() {}

func errMock(msg string) error {
	return fmt.Errorf("%s", msg)
}
