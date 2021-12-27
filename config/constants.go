package config

import "fmt"

type Config struct {
	ApiKey          string
	Symbol          string
	Strike          float32
	Expiration      string
	PutCall         string
	SlackWebhookUrl string
}

func (c *Config) StrikeFormatted() string {
	return fmt.Sprintf("%.1f", c.Strike)
}
