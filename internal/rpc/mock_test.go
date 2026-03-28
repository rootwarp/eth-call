package rpc

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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

// revertError simulates the error type returned by ethclient when a contract reverts.
// It implements rpc.Error (ErrorCode) and rpc.DataError (ErrorData) as expected by
// ethclient.RevertErrorData: code must be 3 and data must be a hex string.
type revertError struct {
	reason  string
	hexData string
}

func (e *revertError) Error() string {
	return fmt.Sprintf("execution reverted: %s", e.reason)
}

func (e *revertError) ErrorCode() int {
	return 3 // EVM revert error code
}

func (e *revertError) ErrorData() interface{} {
	return e.hexData
}

// newRevertError creates a revert error with ABI-encoded Error(string) data.
// The data is hex-encoded as expected by ethclient.RevertErrorData.
func newRevertError(reason string) *revertError {
	// ABI encode: Error(string) selector + encoded string
	selector := crypto.Keccak256([]byte("Error(string)"))[:4]

	// ABI-encoded string: offset (32) + length (32) + padded data
	offset := make([]byte, 32)
	binary.BigEndian.PutUint64(offset[24:], 32)

	length := make([]byte, 32)
	binary.BigEndian.PutUint64(length[24:], uint64(len(reason)))

	padded := make([]byte, ((len(reason)+31)/32)*32)
	copy(padded, reason)

	data := make([]byte, 0, len(selector)+len(offset)+len(length)+len(padded))
	data = append(data, selector...)
	data = append(data, offset...)
	data = append(data, length...)
	data = append(data, padded...)

	return &revertError{
		reason:  reason,
		hexData: "0x" + hex.EncodeToString(data),
	}
}
