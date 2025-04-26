package main

import (
	_ "embed"
	"log/slog"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/metarsit/riskiiit/internal/web3/riskiiit"
)

const (
	wsURL = "wss://api.mainnet.abs.xyz/ws"
)

var (
	riskiiitAddress = common.HexToAddress("0xB4b55C656c6b89f020a6E1044B66D227B638C474")

	//go:embed internal/web3/riskiiit/abi.json
	abiJSON string
)

func main() {
	wsClient, err := riskiiit.NewWSClient(wsURL, abiJSON)
	if err != nil {
		slog.Error("Failed to create WS client", "error", err)
		os.Exit(1)
	}
	defer wsClient.Close()

	if err := wsClient.Subscribe(); err != nil {
		slog.Error("Failed to subscribe to events", "error", err)
		os.Exit(1)
	}
}
