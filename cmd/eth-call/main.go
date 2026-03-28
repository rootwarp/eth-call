// Package main is the entrypoint for the eth-call CLI tool.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli/v2"

	internalabi "github.com/rootwarp/eth-call/internal/abi"
	"github.com/rootwarp/eth-call/internal/converter"
	"github.com/rootwarp/eth-call/internal/encoder"
	"github.com/rootwarp/eth-call/internal/rpc"
	"github.com/rootwarp/eth-call/internal/txbuilder"
)

// dialFn is the function used to connect to an RPC endpoint.
// Override in tests to inject a mock client.
var dialFn = func(ctx context.Context, url string) (rpc.Client, error) {
	return rpc.Dial(ctx, url)
}

func buildApp() *cli.App {
	return &cli.App{
		Name:      "eth-call",
		Usage:     "Generate unsigned Ethereum transactions from ABI",
		UsageText: "eth-call --abi <path> --to <address> [flags] <method> [args...]",
		Version:   "0.1.0",
		Description: `Examples:
  # ERC-20 transfer
  eth-call --abi erc20.json --to 0xdAC17F958D2ee523a2206206994597C13D831ec7 transfer 0xRecipient 1000000

  # Query balance (calldata only)
  eth-call --abi erc20.json --to 0xdAC17F958D2ee523a2206206994597C13D831ec7 --calldata-only balanceOf 0xHolder

  # Uniswap swap with RPC
  eth-call --abi router.json --to 0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D --rpc http://localhost:8545 swapExactTokensForTokens 1000 0 '["0xA","0xB"]' 0xRecipient 9999999999`,
		Before: func(c *cli.Context) error {
			addr := c.String("to")
			if addr != "" && !common.IsHexAddress(addr) {
				return fmt.Errorf("invalid address: %s (expected 0x-prefixed 40-character hex)", addr)
			}
			fromAddr := c.String("from")
			if fromAddr != "" && !common.IsHexAddress(fromAddr) {
				return fmt.Errorf("invalid --from address: %s (expected 0x-prefixed 40-character hex)", fromAddr)
			}
			return nil
		},
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:     "abi",
				Usage:    "path to JSON ABI file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "to",
				Usage:    "target contract address (0x-prefixed hex)",
				Required: true,
			},
			&cli.Int64Flag{
				Name:  "chain-id",
				Value: 1,
				Usage: "chain ID for EIP-155 replay protection",
			},
			&cli.StringFlag{
				Name:  "value",
				Value: "0",
				Usage: "ETH value in wei",
			},
			&cli.BoolFlag{
				Name:  "calldata-only",
				Usage: "output only the ABI-encoded calldata",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "output transaction fields as pretty-printed JSON",
			},
			&cli.StringFlag{
				Name:  "from",
				Usage: "sender address for RPC calls (0x-prefixed hex)",
			},
			&cli.StringFlag{
				Name:    "rpc",
				Usage:   "Ethereum JSON-RPC endpoint URL",
				EnvVars: []string{"ETH_RPC_URL"},
			},
		},
		Action: func(c *cli.Context) error {
			// 1. Parse flags
			abiPath := c.Path("abi")
			toAddr := common.HexToAddress(c.String("to"))
			method := c.Args().First()
			rawArgs := c.Args().Tail()

			// 2. Load ABI
			parsedABI, err := internalabi.LoadFromFile(abiPath)
			if err != nil {
				return err
			}

			// 3. If no method name, list available methods and exit
			if method == "" {
				methods := internalabi.ListMethods(parsedABI)
				_, _ = fmt.Fprintln(c.App.ErrWriter, "Available methods:")
				for _, m := range methods {
					_, _ = fmt.Fprintf(c.App.ErrWriter, "  %s\n", m)
				}
				return nil
			}

			// 4. Look up method
			methodDef, err := internalabi.FindMethod(parsedABI, method)
			if err != nil {
				return err
			}

			// 5. Convert arguments
			args, err := converter.ConvertArgs(rawArgs, methodDef.Inputs)
			if err != nil {
				return err
			}

			// 6a. Calldata-only mode
			if c.Bool("calldata-only") {
				calldataHex, encErr := encoder.EncodeToHex(parsedABI, method, args)
				if encErr != nil {
					return encErr
				}
				_, _ = fmt.Fprintln(c.App.Writer, calldataHex)
				return nil
			}

			// 6b. Validate --rpc requires --from
			rpcURL := c.String("rpc")
			if rpcURL != "" && c.String("from") == "" {
				return fmt.Errorf("--from is required when --rpc is provided")
			}

			// 6c. Encode calldata
			calldata, err := encoder.Encode(parsedABI, method, args)
			if err != nil {
				return err
			}

			// 7. Parse --value
			value, err := parseValue(c.String("value"))
			if err != nil {
				return err
			}

			// 8. Build base params from CLI flags
			params := txbuilder.TxParams{
				ChainID: big.NewInt(c.Int64("chain-id")),
				Value:   value,
			}

			// 9. If --rpc is set, fetch params from RPC and merge
			if rpcURL != "" {
				fromAddr := common.HexToAddress(c.String("from"))

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				client, dialErr := dialFn(ctx, rpcURL)
				if dialErr != nil {
					return dialErr
				}
				defer client.Close()

				rpcParams, fetchErr := client.FetchParams(ctx, fromAddr, toAddr, calldata, value)
				if fetchErr != nil {
					return fetchErr
				}

				// Merge: CLI flags take precedence over RPC values
				params = mergeParams(c, rpcParams, value)
			}

			// 10. Build transaction
			if c.Bool("json") {
				tx, buildErr := txbuilder.BuildTx(calldata, toAddr, params)
				if buildErr != nil {
					return buildErr
				}
				jsonOut, fmtErr := formatTxJSON(tx)
				if fmtErr != nil {
					return fmtErr
				}
				_, _ = fmt.Fprintln(c.App.Writer, jsonOut)
				return nil
			}

			txHex, err := txbuilder.Build(calldata, toAddr, params)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(c.App.Writer, txHex)
			return nil
		},
	}
}

// parseValue converts a decimal or 0x-prefixed hex string to *big.Int.
func parseValue(s string) (*big.Int, error) {
	v := new(big.Int)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		if _, ok := v.SetString(s[2:], 16); !ok {
			return nil, fmt.Errorf("invalid --value: %q (expected decimal or 0x-prefixed hex integer)", s)
		}
		return v, nil
	}
	if _, ok := v.SetString(s, 10); !ok {
		return nil, fmt.Errorf("invalid --value: %q (expected decimal or 0x-prefixed hex integer)", s)
	}
	return v, nil
}

// mergeParams merges RPC-fetched params with CLI-specified overrides.
// CLI flags take precedence when explicitly set by the user.
func mergeParams(c *cli.Context, rpcParams txbuilder.TxParams, cliValue *big.Int) txbuilder.TxParams {
	params := rpcParams

	if c.IsSet("chain-id") {
		params.ChainID = big.NewInt(c.Int64("chain-id"))
	}
	if c.IsSet("value") {
		params.Value = cliValue
	}

	return params
}

// txJSON is the structured output format for --json mode.
type txJSON struct {
	Type                 string `json:"type"`
	ChainID              string `json:"chainId"`
	Nonce                string `json:"nonce"`
	To                   string `json:"to"`
	Value                string `json:"value"`
	Gas                  string `json:"gas"`
	MaxFeePerGas         string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
	Data                 string `json:"data"`
	Raw                  string `json:"raw"`
}

// formatTxJSON formats a transaction as pretty-printed JSON with hex-encoded fields.
func formatTxJSON(tx *types.Transaction) (string, error) {
	raw, err := tx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("failed to serialize transaction: %w", err)
	}

	out := txJSON{
		Type:                 fmt.Sprintf("0x%x", tx.Type()),
		ChainID:              bigToHex(tx.ChainId()),
		Nonce:                fmt.Sprintf("0x%x", tx.Nonce()),
		To:                   strings.ToLower(tx.To().Hex()),
		Value:                bigToHex(tx.Value()),
		Gas:                  fmt.Sprintf("0x%x", tx.Gas()),
		MaxFeePerGas:         bigToHex(tx.GasFeeCap()),
		MaxPriorityFeePerGas: bigToHex(tx.GasTipCap()),
		Data:                 fmt.Sprintf("0x%x", tx.Data()),
		Raw:                  fmt.Sprintf("0x%x", raw),
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(b), nil
}

// bigToHex converts a *big.Int to a 0x-prefixed hex string.
func bigToHex(n *big.Int) string {
	if n == nil || n.Sign() == 0 {
		return "0x0"
	}
	return fmt.Sprintf("0x%x", n)
}

func main() {
	app := buildApp()
	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
