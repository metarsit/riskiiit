package metrics

import "github.com/metarsit/riskiiit/internal/web3/riskiiit"

type Metrics struct {
	score *Score
}

type Score struct {
	// Colors
	Red   int
	Black int
	Green int

	// Winners
	House  int
	Player int

	// TODO: We can also track amount
}

func NewMetrics() (*Metrics, error) {
	return &Metrics{
		score: &Score{
			Red:   0,
			Black: 0,
			Green: 0,

			House:  0,
			Player: 0,
		},
	}, nil
}

func (m *Metrics) Collect(c chan riskiiit.SpinResolvedEvent) {
	for {
		select {
		case event := <-c:
			// Winners
			if event.Result {
				m.score.Player++
			} else {
				m.score.House++
			}

			// Colors
			switch event.ActualResult {
			case riskiiit.Red:
				m.score.Red++
			case riskiiit.Black:
				m.score.Black++
			case riskiiit.Green:
				m.score.Green++
			}
		}
	}
}

func (m *Metrics) GetScore() *Score {
	return m.score
}
