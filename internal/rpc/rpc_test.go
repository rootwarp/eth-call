package rpc

import (
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

// TestFetchParamsPopulatesTxParams verifies that FetchParams returns a correctly populated TxParams.
func TestFetchParamsPopulatesTxParams(t *testing.T) {
	baseFee := big.NewInt(10_000_000_000)
	gasTipCap := big.NewInt(1_000_000_000)
	expectedFeeCap := big.NewInt(21_000_000_000)
	estimatedGas := uint64(50000)
	expectedGas := uint64(60000) // 50000 * 120 / 100

	mock := &mockBackend{
		chainID:     big.NewInt(1),
		nonce:       42,
		gasTipCap:   gasTipCap,
		baseFee:     baseFee,
		gasEstimate: estimatedGas,
	}

	ec := &ethClient{backend: mock}

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
}

// TestFetchParamsChainIDError verifies error wrapping when ChainID fails.
func TestFetchParamsChainIDError(t *testing.T) {
	mock := &mockBackend{
		chainIDErr: errMock("chain id unavailable"),
	}

	ec := &ethClient{backend: mock}
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	_, err := ec.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "rpc: failed to fetch chain ID") {
		t.Fatalf("expected error containing %q, got %q", "rpc: failed to fetch chain ID", err.Error())
	}
}

// TestFetchParamsNonceError verifies error wrapping when nonce fetch fails.
func TestFetchParamsNonceError(t *testing.T) {
	mock := &mockBackend{
		chainID:  big.NewInt(1),
		nonceErr: errMock("nonce unavailable"),
	}

	ec := &ethClient{backend: mock}
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	_, err := ec.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "rpc: failed to fetch nonce") {
		t.Fatalf("expected error containing %q, got %q", "rpc: failed to fetch nonce", err.Error())
	}
}

// TestFetchParamsGasTipCapError verifies error wrapping when gas tip cap fetch fails.
func TestFetchParamsGasTipCapError(t *testing.T) {
	mock := &mockBackend{
		chainID:      big.NewInt(1),
		nonce:        1,
		gasTipCapErr: errMock("tip cap unavailable"),
	}

	ec := &ethClient{backend: mock}
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	_, err := ec.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "rpc: failed to fetch gas tip cap") {
		t.Fatalf("expected error containing %q, got %q", "rpc: failed to fetch gas tip cap", err.Error())
	}
}

// TestFetchParamsHeaderError verifies error wrapping when header fetch fails.
func TestFetchParamsHeaderError(t *testing.T) {
	mock := &mockBackend{
		chainID:   big.NewInt(1),
		nonce:     1,
		gasTipCap: big.NewInt(1_000_000_000),
		headerErr: errMock("header unavailable"),
	}

	ec := &ethClient{backend: mock}
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	_, err := ec.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "rpc: failed to fetch base fee") {
		t.Fatalf("expected error containing %q, got %q", "rpc: failed to fetch base fee", err.Error())
	}
}

// TestFetchParamsEstimateGasError verifies error wrapping when gas estimation fails.
func TestFetchParamsEstimateGasError(t *testing.T) {
	mock := &mockBackend{
		chainID:        big.NewInt(1),
		nonce:          1,
		gasTipCap:      big.NewInt(1_000_000_000),
		baseFee:        big.NewInt(10_000_000_000),
		gasEstimateErr: errMock("gas estimation failed"),
	}

	ec := &ethClient{backend: mock}
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	_, err := ec.FetchParams(context.Background(), from, to, nil, big.NewInt(0))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "rpc: failed to estimate gas") {
		t.Fatalf("expected error containing %q, got %q", "rpc: failed to estimate gas", err.Error())
	}
}

// TestFetchParamsReturnsCorrectValuePassthrough verifies the value is passed through to TxParams.
func TestFetchParamsReturnsCorrectValuePassthrough(t *testing.T) {
	oneETH := new(big.Int).SetUint64(1_000_000_000_000_000_000)

	mock := &mockBackend{
		chainID:     big.NewInt(1),
		nonce:       0,
		gasTipCap:   big.NewInt(0),
		baseFee:     big.NewInt(0),
		gasEstimate: 21000,
	}

	ec := &ethClient{backend: mock}
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	params, err := ec.FetchParams(context.Background(), from, to, nil, oneETH)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertBigInt(t, "Value", params.Value, oneETH)
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
