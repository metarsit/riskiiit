package riskiiit

import (
	"encoding/json"
	"math/big"
)

const (
	Red uint8 = iota
	Black
	Green
)

type SpinResolvedEvent struct {
	Result       bool
	PlayerChoice uint8
	ActualResult uint8
	BetAmount    *big.Int
	Payout       *big.Int
	WinStreak    *big.Int
	Username     string
}

func (s *SpinResolvedEvent) String() string {
	b, err := json.Marshal(s)
	if err != nil {
		return ""
	}

	return string(b)
}
