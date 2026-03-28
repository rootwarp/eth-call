// Package abi loads Solidity JSON ABI files and resolves method signatures.
package abi

import (
	"fmt"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

// LoadFromFile reads and parses a JSON ABI file.
func LoadFromFile(path string) (ethabi.ABI, error) {
	_ = path
	return ethabi.ABI{}, fmt.Errorf("not implemented")
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
