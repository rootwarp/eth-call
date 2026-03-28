package abi

import (
	"fmt"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

// LoadFromFile reads and parses a JSON ABI file.
func LoadFromFile(path string) (ethabi.ABI, error) {
	return ethabi.ABI{}, fmt.Errorf("not implemented")
}

// FindMethod looks up a method by name in the parsed ABI.
func FindMethod(parsedABI ethabi.ABI, name string) (ethabi.Method, error) {
	return ethabi.Method{}, fmt.Errorf("not implemented")
}

// ListMethods returns a sorted list of method signatures from the ABI.
func ListMethods(parsedABI ethabi.ABI) []string {
	return nil
}
