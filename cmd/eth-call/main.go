// Package main is the entrypoint for the eth-call CLI tool.
package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func buildApp() *cli.App {
	return &cli.App{
		Name:      "eth-call",
		Usage:     "Generate unsigned Ethereum transactions from ABI",
		UsageText: "eth-call --abi <path> --to <address> [flags] <method> [args...]",
		Version:   "0.1.0",
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
