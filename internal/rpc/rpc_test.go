package rpc

import (
	"bytes"
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rootwarp/eth-call/internal/txbuilder"
)

// TestClientInterfaceCompliance verifies that ethClient implements Client.
func TestClientInterfaceCompliance(_ *testing.T) {
	var _ Client = (*ethClient)(nil)
}

// TestMockClientInterfaceCompliance verifies that MockClient implements Client.
func TestMockClientInterfaceCompliance(_ *testing.T) {
	var _ Client = (*MockClient)(nil)
}

// TestDialInvalidURL verifies Dial returns an error for invalid URLs.
func TestDialInvalidURL(t *testing.T) {
	ctx := context.Background()

	_, err := Dial(ctx, "://invalid")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

// TestDialEmptyURL verifies Dial returns an error for empty URL.
func TestDialEmptyURL(t *testing.T) {
	ctx := context.Background()

	_, err := Dial(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

// TestDialReturnsClient verifies Dial returns a non-nil Client for valid-looking URLs.
func TestDialReturnsClient(t *testing.T) {
	ctx := context.Background()

	client, err := Dial(ctx, "http://localhost:8545")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	client.Close()
}

// TestComputeGasFeeCap verifies gasFeeCap = 2*baseFee + gasTipCap.
func TestComputeGasFeeCap(t *testing.T) {
	tests := []struct {
		want      *big.Int
		baseFee   *big.Int
		gasTipCap *big.Int
		name      string
	}{
		{
			name:      "basic computation",
			baseFee:   big.NewInt(10_000_000_000),
			gasTipCap: big.NewInt(1_000_000_000),
			want:      big.NewInt(21_000_000_000),
		},
		{
			name:      "zero base fee",
			baseFee:   big.NewInt(0),
			gasTipCap: big.NewInt(2_000_000_000),
			want:      big.NewInt(2_000_000_000),
		},
		{
			name:      "zero tip",
			baseFee:   big.NewInt(5_000_000_000),
			gasTipCap: big.NewInt(0),
			want:      big.NewInt(10_000_000_000),
		},
		{
			name:      "both zero",
			baseFee:   big.NewInt(0),
			gasTipCap: big.NewInt(0),
			want:      big.NewInt(0),
		},
		{
			name:      "large values",
			baseFee:   big.NewInt(100_000_000_000),
			gasTipCap: big.NewInt(5_000_000_000),
			want:      big.NewInt(205_000_000_000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeGasFeeCap(tt.baseFee, tt.gasTipCap)
			if got.Cmp(tt.want) != 0 {
				t.Fatalf("computeGasFeeCap(%s, %s) = %s, want %s", tt.baseFee, tt.gasTipCap, got, tt.want)
			}
		})
	}
}

// TestApplyGasMargin verifies the 20% safety margin on gas limit.
func TestApplyGasMargin(t *testing.T) {
	tests := []struct {
		name     string
		gasLimit uint64
		want     uint64
	}{
		{"basic", 100000, 120000},
		{"zero", 0, 0},
		{"small", 100, 120},
		{"standard transfer", 21000, 25200},
		{"odd value", 33333, 39999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyGasMargin(tt.gasLimit)
			if got != tt.want {
				t.Fatalf("applyGasMargin(%d) = %d, want %d", tt.gasLimit, got, tt.want)
			}
		})
	}
}

// --- FetchParams tests with graceful degradation ---

// TestFetchParamsAllSucceed verifies FetchParams returns fully populated TxParams when all RPC calls succeed.
func TestFetchParamsAllSucceed(t *testing.T) {
	baseFee := big.NewInt(10_000_000_000)
	gasTipCap := big.NewInt(1_000_000_000)
	expectedFeeCap := big.NewInt(21_000_000_000)
	estimatedGas := uint64(50000)
	expectedGas := uint64(60000) // 50000 * 120 / 100

	var warnBuf bytes.Buffer
	ec := &ethClient{
		backend: &mockBackend{
			chainID:     big.NewInt(1),
			nonce:       42,
			gasTipCap:   gasTipCap,
			baseFee:     baseFee,
			gasEstimate: estimatedGas,
		},
		warnWriter: &warnBuf,
	}

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")
	calldata := []byte{0xa9, 0x05, 0x9c, 0xbb}
	value := big.NewInt(0)

	params, err := ec.FetchParams(context.Background(), from, to, calldata, value)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertBigInt(t, "ChainID", params.ChainID, big.NewInt(1))
	if params.Nonce != 42 {
		t.Fatalf("expected nonce 42, got %d", params.Nonce)
	}
	assertBigInt(t, "GasTipCap", params.GasTipCap, gasTipCap)
	assertBigInt(t, "GasFeeCap", params.GasFeeCap, expectedFeeCap)
	if params.GasLimit != expectedGas {
		t.Fatalf("expected gas limit %d, got %d", expectedGas, params.GasLimit)
	}
	assertBigInt(t, "Value", params.Value, value)

	if warnBuf.Len() != 0 {
		t.Fatalf("expected no warnings, got: %s", warnBuf.String())
	}
}

// TestFetchParamsChainIDDegrades verifies ChainID failure produces warning + nil default.
func TestFetchParamsChainIDDegrades(t *testing.T) {
	var warnBuf bytes.Buffer
	ec := &ethClient{
		backend: &mockBackend{
			chainIDErr:  errMock("chain id unavailable"),
			nonce:       1,
			gasTipCap:   big.NewInt(1_000_000_000),
			baseFee:     big.NewInt(10_000_000_000),
			gasEstimate: 21000,
		},
		warnWriter: &warnBuf,
	}

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	params, err := ec.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err != nil {
		t.Fatalf("expected no error on degraded chain ID, got: %v", err)
	}

	if params.ChainID != nil {
		t.Fatalf("expected nil ChainID on degradation, got %s", params.ChainID)
	}

	if !strings.Contains(warnBuf.String(), "chain ID") {
		t.Fatalf("expected warning about chain ID, got: %s", warnBuf.String())
	}
}

// TestFetchParamsNonceError verifies nonce failure is NOT degradable — always a hard error.
func TestFetchParamsNonceError(t *testing.T) {
	var warnBuf bytes.Buffer
	ec := &ethClient{
		backend: &mockBackend{
			chainID:  big.NewInt(1),
			nonceErr: errMock("nonce unavailable"),
		},
		warnWriter: &warnBuf,
	}

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	_, err := ec.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err == nil {
		t.Fatal("expected error for nonce failure, got nil")
	}

	if !strings.Contains(err.Error(), "rpc: failed to fetch nonce") {
		t.Fatalf("expected error containing %q, got %q", "rpc: failed to fetch nonce", err.Error())
	}
}

// TestFetchParamsGasTipCapDegrades verifies gas tip cap failure produces warning + zero default.
func TestFetchParamsGasTipCapDegrades(t *testing.T) {
	var warnBuf bytes.Buffer
	ec := &ethClient{
		backend: &mockBackend{
			chainID:      big.NewInt(1),
			nonce:        1,
			gasTipCapErr: errMock("tip cap unavailable"),
			baseFee:      big.NewInt(10_000_000_000),
			gasEstimate:  21000,
		},
		warnWriter: &warnBuf,
	}

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	params, err := ec.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err != nil {
		t.Fatalf("expected no error on degraded gas tip cap, got: %v", err)
	}

	assertBigInt(t, "GasTipCap", params.GasTipCap, big.NewInt(0))
	// GasFeeCap = 2*baseFee + 0 = 20 gwei
	assertBigInt(t, "GasFeeCap", params.GasFeeCap, big.NewInt(20_000_000_000))

	if !strings.Contains(warnBuf.String(), "gas tip cap") {
		t.Fatalf("expected warning about gas tip cap, got: %s", warnBuf.String())
	}
}

// TestFetchParamsHeaderDegrades verifies header failure produces warning + zero gas fee cap.
func TestFetchParamsHeaderDegrades(t *testing.T) {
	var warnBuf bytes.Buffer
	ec := &ethClient{
		backend: &mockBackend{
			chainID:     big.NewInt(1),
			nonce:       1,
			gasTipCap:   big.NewInt(1_000_000_000),
			headerErr:   errMock("header unavailable"),
			gasEstimate: 21000,
		},
		warnWriter: &warnBuf,
	}

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	params, err := ec.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err != nil {
		t.Fatalf("expected no error on degraded header, got: %v", err)
	}

	assertBigInt(t, "GasFeeCap", params.GasFeeCap, big.NewInt(0))
	assertBigInt(t, "GasTipCap", params.GasTipCap, big.NewInt(1_000_000_000))

	if !strings.Contains(warnBuf.String(), "base fee") {
		t.Fatalf("expected warning about base fee, got: %s", warnBuf.String())
	}
}

// TestFetchParamsEstimateGasDegrades verifies gas estimation failure produces warning + zero default.
func TestFetchParamsEstimateGasDegrades(t *testing.T) {
	var warnBuf bytes.Buffer
	ec := &ethClient{
		backend: &mockBackend{
			chainID:        big.NewInt(1),
			nonce:          1,
			gasTipCap:      big.NewInt(1_000_000_000),
			baseFee:        big.NewInt(10_000_000_000),
			gasEstimateErr: errMock("gas estimation failed"),
		},
		warnWriter: &warnBuf,
	}

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	params, err := ec.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err != nil {
		t.Fatalf("expected no error on degraded gas estimate, got: %v", err)
	}

	if params.GasLimit != 0 {
		t.Fatalf("expected gas limit 0, got %d", params.GasLimit)
	}

	if !strings.Contains(warnBuf.String(), "estimate gas") {
		t.Fatalf("expected warning about gas estimation, got: %s", warnBuf.String())
	}
}

// TestFetchParamsEstimateGasRevert verifies revert reason is extracted from EstimateGas failure.
func TestFetchParamsEstimateGasRevert(t *testing.T) {
	revertErr := newRevertError("insufficient balance")

	var warnBuf bytes.Buffer
	ec := &ethClient{
		backend: &mockBackend{
			chainID:        big.NewInt(1),
			nonce:          1,
			gasTipCap:      big.NewInt(1_000_000_000),
			baseFee:        big.NewInt(10_000_000_000),
			gasEstimateErr: revertErr,
		},
		warnWriter: &warnBuf,
	}

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	params, err := ec.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err != nil {
		t.Fatalf("expected no error on degraded gas estimate with revert, got: %v", err)
	}

	if params.GasLimit != 0 {
		t.Fatalf("expected gas limit 0, got %d", params.GasLimit)
	}

	warnStr := warnBuf.String()
	if !strings.Contains(warnStr, "insufficient balance") {
		t.Fatalf("expected warning to contain revert reason, got: %s", warnStr)
	}
}

// TestFetchParamsEstimateGasRevertBadData verifies that invalid revert data falls back to generic warning.
func TestFetchParamsEstimateGasRevertBadData(t *testing.T) {
	// An error with ErrorData() but invalid ABI encoding
	badRevert := &revertError{
		reason:  "bad",
		hexData: "0xdead", // too short to be valid ABI
	}

	var warnBuf bytes.Buffer
	ec := &ethClient{
		backend: &mockBackend{
			chainID:        big.NewInt(1),
			nonce:          1,
			gasTipCap:      big.NewInt(1_000_000_000),
			baseFee:        big.NewInt(10_000_000_000),
			gasEstimateErr: badRevert,
		},
		warnWriter: &warnBuf,
	}

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	params, err := ec.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if params.GasLimit != 0 {
		t.Fatalf("expected gas limit 0, got %d", params.GasLimit)
	}

	// Should fall back to generic warning (not revert reason)
	warnStr := warnBuf.String()
	if !strings.Contains(warnStr, "estimate gas") {
		t.Fatalf("expected warning about gas estimation, got: %s", warnStr)
	}
	if strings.Contains(warnStr, "contract reverted") {
		t.Fatalf("did not expect revert reason in warning for bad data, got: %s", warnStr)
	}
}

// TestFetchParamsMultipleDegradations verifies multiple degradations work together.
func TestFetchParamsMultipleDegradations(t *testing.T) {
	var warnBuf bytes.Buffer
	ec := &ethClient{
		backend: &mockBackend{
			chainIDErr:     errMock("chain id unavailable"),
			nonce:          5,
			gasTipCapErr:   errMock("tip cap unavailable"),
			headerErr:      errMock("header unavailable"),
			gasEstimateErr: errMock("gas estimation failed"),
		},
		warnWriter: &warnBuf,
	}

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	params, err := ec.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err != nil {
		t.Fatalf("expected no error with multiple degradations, got: %v", err)
	}

	if params.ChainID != nil {
		t.Fatalf("expected nil ChainID, got %s", params.ChainID)
	}
	assertBigInt(t, "GasTipCap", params.GasTipCap, big.NewInt(0))
	assertBigInt(t, "GasFeeCap", params.GasFeeCap, big.NewInt(0))
	if params.GasLimit != 0 {
		t.Fatalf("expected gas limit 0, got %d", params.GasLimit)
	}
	if params.Nonce != 5 {
		t.Fatalf("expected nonce 5, got %d", params.Nonce)
	}

	warnStr := warnBuf.String()
	warnLines := strings.Count(warnStr, "warning:")
	if warnLines != 4 {
		t.Fatalf("expected 4 warnings, got %d: %s", warnLines, warnStr)
	}
}

// TestFetchParamsValuePassthrough verifies the value is passed through to TxParams.
func TestFetchParamsValuePassthrough(t *testing.T) {
	oneETH := new(big.Int).SetUint64(1_000_000_000_000_000_000)

	var warnBuf bytes.Buffer
	ec := &ethClient{
		backend: &mockBackend{
			chainID:     big.NewInt(1),
			gasTipCap:   big.NewInt(0),
			baseFee:     big.NewInt(0),
			gasEstimate: 21000,
		},
		warnWriter: &warnBuf,
	}

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	params, err := ec.FetchParams(context.Background(), from, to, nil, oneETH)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertBigInt(t, "Value", params.Value, oneETH)
}

// --- MockClient tests ---

// TestMockClientFetchParams verifies the MockClient works correctly.
func TestMockClientFetchParams(t *testing.T) {
	mock := &MockClient{
		FetchParamsFn: func(_ context.Context, _, _ common.Address, _ []byte, _ *big.Int) (txbuilder.TxParams, error) {
			return txbuilder.TxParams{
				ChainID:   big.NewInt(42),
				Nonce:     10,
				GasTipCap: big.NewInt(1_000_000_000),
				GasFeeCap: big.NewInt(20_000_000_000),
				GasLimit:  50000,
				Value:     big.NewInt(0),
			}, nil
		},
	}

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	params, err := mock.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertBigInt(t, "ChainID", params.ChainID, big.NewInt(42))
	if params.Nonce != 10 {
		t.Fatalf("expected nonce 10, got %d", params.Nonce)
	}

	mock.Close()
}

// TestMockClientDefaultFetchParams verifies MockClient returns zero TxParams when no FetchParamsFn set.
func TestMockClientDefaultFetchParams(t *testing.T) {
	mock := &MockClient{}

	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	params, err := mock.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params.ChainID != nil {
		t.Fatalf("expected nil ChainID, got %s", params.ChainID)
	}
}

// TestMockClientCloseFn verifies MockClient calls CloseFn when set.
func TestMockClientCloseFn(t *testing.T) {
	called := false
	mock := &MockClient{
		CloseFn: func() { called = true },
	}

	mock.Close()

	if !called {
		t.Fatal("expected CloseFn to be called")
	}
}

// --- Test helpers ---

func assertBigInt(t *testing.T, field string, got, want *big.Int) {
	t.Helper()
	if got.Cmp(want) != 0 {
		t.Fatalf("expected %s %s, got %s", field, want, got)
	}
}

// TestTxParamsFieldTypes verifies TxParams field types at compile time.
func TestTxParamsFieldTypes(_ *testing.T) {
	_ = txbuilder.TxParams{
		ChainID:   big.NewInt(1),
		Nonce:     uint64(0),
		GasTipCap: big.NewInt(0),
		GasFeeCap: big.NewInt(0),
		GasLimit:  uint64(0),
		Value:     big.NewInt(0),
	}
}
