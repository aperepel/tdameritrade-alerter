package chain

import (
	"fmt"
	"github.com/slack-go/slack"
	"strings"
	"tda-watch/config"
)

type SlackProcessor struct {
	Config config.Config
}

func (s *SlackProcessor) Name() string {
	return "SlackNotifier"
}

func (s *SlackProcessor) Analyze(optionChains *Chains) error {
	slackURL := s.Config.SlackWebhookUrl
	if slackURL == "" {
		return fmt.Errorf("slack webhook URL not configured")
	}

	requestedStrike := fmt.Sprintf("%.1f", s.Config.Strike)

	builder := strings.Builder{}
	fmt.Fprintf(&builder, "*%s* %s %s",
		optionChains.Symbol,
		requestedStrike,
		s.Config.Expiration,
	)

	// initialize fields with basic info
	instrumentFields := []slack.AttachmentField{
		{
			Title: "Underlying Price",
			Value: fmt.Sprintf("%.2f\n", optionChains.UnderlyingPrice),
			Short: true,
		},
		{
			Title: "Delayed",
			Value: fmt.Sprintf("%t", optionChains.IsDelayed),
			Short: true,
		},
	}

	// fill out the values we have now, append more values below in the options loop
	optionFields := []slack.AttachmentField{
		{
			Title: "Strike",
			Value: requestedStrike,
			Short: true,
		},
	}

	expDateMap := optionChains.CallExpDateMap
	for expiration, strikeMap := range expDateMap {
		// response value will have e.g. '2021-12-31:5' drop everything after the ':'
		cleansedExp := strings.Split(expiration, ":")[0]
		if s.Config.Expiration == cleansedExp {
			// found the requested expiration
			options := strikeMap[requestedStrike]
			for _, option := range options {
				optionFields = append(optionFields, slack.AttachmentField{
					Title: "DTE",
					Value: fmt.Sprintf("%d", option.DaysToExpiration),
					Short: true,
				})
				optionFields = append(optionFields, slack.AttachmentField{
					Title: "Delta",
					Value: fmt.Sprintf("%.2f", option.Delta),
					Short: true,
				})
				optionFields = append(optionFields, slack.AttachmentField{
					Title: "Bid/Ask",
					Value: fmt.Sprintf("%.2f/%.2f", option.Bid, option.Ask),
					Short: true,
				})
			}
		}
	}

	title := fmt.Sprintf("%s %s %s",
		optionChains.Symbol,
		requestedStrike,
		s.Config.Expiration)

	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{
			{
				Color:      "info",
				Title:      title,
				Fields:     instrumentFields,
				MarkdownIn: []string{"fields"},
			},
			{
				Color:      "info",
				Title:      "     ---=====  Option Details =====---  ",
				Fields:     optionFields,
				MarkdownIn: []string{"fields"},
			},
		},
	}

	return slack.PostWebhook(s.Config.SlackWebhookUrl, &msg)
}
