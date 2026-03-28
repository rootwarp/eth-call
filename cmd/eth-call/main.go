// Package main is the entrypoint for the eth-call CLI tool.
package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
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
			_ = c
			return fmt.Errorf("not implemented")
		},
	}
}

func main() {
	app := buildApp()
	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
