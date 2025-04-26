package main

import (
	_ "embed"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/metarsit/riskiiit/internal/metrics"
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
	c := make(chan riskiiit.SpinResolvedEvent)
	defer close(c)

	// Initialization
	wsClient, err := riskiiit.NewWSClient(wsURL, abiJSON)
	if err != nil {
		slog.Error("Failed to create WS client", "error", err)
		os.Exit(1)
	}
	defer wsClient.Close()

	metrics, err := metrics.NewMetrics()
	if err != nil {
		slog.Error("Failed to create metrics", "error", err)
		os.Exit(1)
	}

	// Application
	slog.Info("Subscribing to events")
	go wsClient.Subscribe(c)
	slog.Info("Collecting metrics")
	go metrics.Collect(c)

	go func() {
		// every 2 seconds print the metrics
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				data := metrics.GetScore()
				slog.Info("Metrics", "red", data.Red, "black", data.Black, "green", data.Green, "house", data.House, "player", data.Player)
			}
		}
	}()

	// Block exit
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	<-exit

	slog.Info("Exiting")
}
