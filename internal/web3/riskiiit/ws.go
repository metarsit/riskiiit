package riskiiit

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var riskiiitAddress = common.HexToAddress("0xB4b55C656c6b89f020a6E1044B66D227B638C474")

type WSClient struct {
	client    *ethclient.Client
	parsedABI abi.ABI

	wsURL string
}

func NewWSClient(wsURL, abiJSON string) (*WSClient, error) {
	// 1. Initialize the Websocket client
	client, err := ethclient.Dial(wsURL)
	if err != nil {
		return nil, err
	}

	// 2. Parse the ABI (We will to find it on Abscan)
	// https://abscan.org/address/0xebac5872d5d3a53e03c6953bee7584201dd38759#code
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}

	return &WSClient{
		client:    client,
		parsedABI: parsedABI,
		wsURL:     wsURL,
	}, nil
}

func (c *WSClient) Close() {
	c.client.Close()
}

func (c *WSClient) Subscribe(ch chan<- SpinResolvedEvent) error {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{riskiiitAddress},
	}

	logs := make(chan types.Log)
	sub, err := c.client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			slog.Error("Subscription error", "error", err)
		case vLog := <-logs:
			event, err := c.parsedABI.EventByID(vLog.Topics[0])
			if err != nil {
				slog.Error("Failed to get event by ID", "error", err)
				continue
			}

			m := make(map[string]interface{})
			if event.Name != "SpinResolved" {
				slog.Debug("Skipping event", "event", event.Name)
				continue
			}

			if err := c.parsedABI.UnpackIntoMap(m, event.Name, vLog.Data); err != nil {
				slog.Error("Failed to unpack into map", "error", err)
				continue
			}

			b, err := json.Marshal(m)
			if err != nil {
				slog.Error("Failed to marshal map", "error", err)
				continue
			}

			var spinEvent SpinResolvedEvent
			if err := json.Unmarshal(b, &spinEvent); err != nil {
				slog.Error("Failed to unmarshal event", "error", err)
				continue
			}

			ch <- spinEvent
		}
	}
}
