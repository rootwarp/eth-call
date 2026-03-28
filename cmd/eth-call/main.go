// Package main is the entrypoint for the eth-call CLI tool.
package main

import (
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"

	internalabi "github.com/rootwarp/eth-call/internal/abi"
	"github.com/rootwarp/eth-call/internal/converter"
	"github.com/rootwarp/eth-call/internal/encoder"
	"github.com/rootwarp/eth-call/internal/txbuilder"
)

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

			// 6b. Encode calldata
			calldata, err := encoder.Encode(parsedABI, method, args)
			if err != nil {
				return err
			}

			// 7. Parse --value
			value, err := parseValue(c.String("value"))
			if err != nil {
				return err
			}

			// 8. Build transaction
			params := txbuilder.TxParams{
				ChainID: big.NewInt(c.Int64("chain-id")),
				Value:   value,
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

func main() {
	app := buildApp()
	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
