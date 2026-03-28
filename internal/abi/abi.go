// Package abi loads Solidity JSON ABI files and resolves method signatures.
package abi

import (
	"fmt"
	"os"
	"sort"
	"strings"

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
// It checks both the resolved name (map key) and the raw name from the ABI.
// If the method is not found, returns an error listing all available methods.
func FindMethod(parsedABI ethabi.ABI, name string) (ethabi.Method, error) {
	// Direct lookup by resolved name (handles exact matches and overloaded names like "foo0")
	if method, ok := parsedABI.Methods[name]; ok {
		return method, nil
	}

	// Fallback: match by RawName for overloaded methods where user provides the base name
	for _, method := range parsedABI.Methods {
		if method.RawName == name {
			return method, nil
		}
	}

	available := strings.Join(ListMethods(parsedABI), ", ")
	return ethabi.Method{}, fmt.Errorf("abi: method %q not found; available methods: %s", name, available)
}

// ListMethods returns a sorted list of method signatures from the ABI.
func ListMethods(parsedABI ethabi.ABI) []string {
	methods := make([]string, 0, len(parsedABI.Methods))
	for _, m := range parsedABI.Methods {
		methods = append(methods, m.Sig)
	}
	sort.Strings(methods)
	return methods
}
