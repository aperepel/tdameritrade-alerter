package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"tdameritrade-alerter/chain"
	"tdameritrade-alerter/util"
)

const (
	baseUrl   = "https://api.tdameritrade.com/v1"
	chainsApi = "/marketdata/chains"
)

func main() {

	if util.IsStandalone() {
		fmt.Println("Running standalone")
	} else {
		log.Info().Msg("Running in a container")
	}

	// executable is the first arg
	if len(os.Args) == 1 {
		log.Fatal().Msg("Please specify the dir with the app config file")
	}

	cfg, err := util.LoadConfig(os.Args[1])
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load the config")
	}

	chainsUrl := fmt.Sprintf(
		"%s%s?apikey=%s&symbol=%s&contractType=%s&strike=%s",
		baseUrl,
		chainsApi,
		url.QueryEscape(cfg.ApiKey),
		url.QueryEscape(cfg.Symbol),
		cfg.PutCall,
		url.QueryEscape(cfg.StrikeFormatted()),
	)

	resp, err := http.Get(chainsUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to invoke the API")
	}

	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read response")
	}
	if resp.StatusCode != 200 {
		log.Fatal().Err(err).Msgf("API call didn't succeed: %q", string(respBytes))
	}
	chains := chain.Chains{}

	var prettyJson bytes.Buffer
	err = json.Indent(&prettyJson, respBytes, "", "  ")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to format response json")
	}
	log.Debug().Msg(prettyJson.String())

	err = json.Unmarshal(respBytes, &chains)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse the json response")
	}

	stdOutProcessor := chain.StdOutProcessor{cfg}
	_ = stdOutProcessor.Analyze(&chains)

	if cfg.SlackWebhookUrl == "" {
		log.Info().Msg("Slack webhook URL not configured")
	} else {
		slackProcessor := chain.SlackProcessor{cfg}
		err := slackProcessor.Analyze(&chains)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to notify via Slack")
		}
	}
}
