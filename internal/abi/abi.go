// Package abi loads Solidity JSON ABI files and resolves method signatures.
package abi

import (
	"fmt"
	"os"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

// LoadFromFile reads and parses a JSON ABI file.
func LoadFromFile(path string) (ethabi.ABI, error) {
	f, err := os.Open(path) //nolint:gosec // path is caller-provided by design
	if err != nil {
		return ethabi.ABI{}, fmt.Errorf("abi: file not found: %w", err)
	}
	defer func() { _ = f.Close() }()

	parsed, err := ethabi.JSON(f)
	if err != nil {
		return ethabi.ABI{}, fmt.Errorf("abi: invalid JSON ABI: %w", err)
	}

	return parsed, nil
}

// FindMethod looks up a method by name in the parsed ABI.
func FindMethod(parsedABI ethabi.ABI, name string) (ethabi.Method, error) {
	_ = parsedABI
	_ = name
	return ethabi.Method{}, fmt.Errorf("not implemented")
}

// ListMethods returns a sorted list of method signatures from the ABI.
func ListMethods(parsedABI ethabi.ABI) []string {
	_ = parsedABI
	return nil
}
