package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"log/slog"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
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
	// 1. Initialize the Websocket client
	client, err := ethclient.Dial(wsURL)
	if err != nil {
		slog.Error("Failed to connect to the Ethereum client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	// 2. Parse the ABI (We will to find it on Abscan)
	// https://abscan.org/address/0xebac5872d5d3a53e03c6953bee7584201dd38759#code
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		slog.Error("Failed to parse the ABI", "error", err)
		os.Exit(1)
	}

	// 3. Filtering the events from Smart Contract Address
	query := ethereum.FilterQuery{
		Addresses: []common.Address{riskiiitAddress},
	}

	// 4. Start subscribing to the events (as well as the socket)
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		slog.Error("Failed to subscribe to the events", "error", err)
		os.Exit(1)
	}
	defer sub.Unsubscribe()

	// 5. Create a function to handle the events
	slog.Info("Starting to listen for events")
	for {
		select {
		case err := <-sub.Err():
			slog.Error("Subscription error", "error", err)
		case vLog := <-logs:
			event, err := parsedABI.EventByID(vLog.Topics[0])
			if err != nil {
				slog.Error("Failed to get event by ID", "error", err)
				continue
			}

			m := make(map[string]interface{})
			if event.Name != "SpinResolved" {
				slog.Debug("Skipping event", "event", event.Name)
				continue
			}

			if err := parsedABI.UnpackIntoMap(m, event.Name, vLog.Data); err != nil {
				slog.Error("Failed to unpack into map", "error", err)
				continue
			}

			b, err := json.Marshal(m)
			if err != nil {
				slog.Error("Failed to marshal map", "error", err)
				continue
			}

			var spinEvent riskiiit.SpinResolvedEvent
			if err := json.Unmarshal(b, &spinEvent); err != nil {
				slog.Error("Failed to unmarshal event", "error", err)
				continue
			}

			slog.Info("Event", "event", spinEvent.String())
		}
	}
}
