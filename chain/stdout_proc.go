package chain

import (
	"fmt"
	"strings"
	"tdameritrade-alerter/config"
)

type StdOutProcessor struct {
	Config config.Config
}

func (s *StdOutProcessor) Name() string {
	return "StdOutPrinter"
}

func (s *StdOutProcessor) Analyze(chains *Chains) error {
	requestedExpiration := s.Config.Expiration
	requestedStrike := s.Config.StrikeFormatted()

	builder := strings.Builder{}

	_, _ = fmt.Fprintf(&builder, "Instrument: %s\n", chains.Symbol)
	_, _ = fmt.Fprintf(&builder, "Underlying price: %.2f\n", chains.UnderlyingPrice)
	_, _ = fmt.Fprintf(&builder, "Delayed: %t\n", chains.IsDelayed)
	_, _ = fmt.Fprintf(&builder, "Option strike: %s\n", requestedStrike)
	_, _ = fmt.Fprintln(&builder)

	expDateMap := chains.CallExpDateMap
	for expiration, strikeMap := range expDateMap {
		// response value will have e.g. '2021-12-31:5' drop everything after the ':'
		cleansedExp := strings.Split(expiration, ":")[0]
		if requestedExpiration == cleansedExp {
			// found the requested expiration
			_, _ = fmt.Fprintf(&builder, "Expiration: %s\n", cleansedExp)

			//fmt.Printf("Len: %d\n", len(strikeMap))
			//fmt.Printf("Strike map: %s\n", strikeMap)
			// uses the decimals in the response
			options := strikeMap[requestedStrike]
			for _, option := range options {
				_, _ = fmt.Fprintf(&builder, "DTE: %.d\n", option.DaysToExpiration)
				_, _ = fmt.Fprintf(&builder, "Delta: %.2f\n", option.Delta)
				_, _ = fmt.Fprintf(&builder, "Bid/ask: %.2f/%.2f", option.Bid, option.Ask)
			}
		}
	}

	output := builder.String()
	fmt.Println(output)

	return nil
}