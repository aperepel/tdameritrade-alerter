package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"tdameritrade-alerter/chain"
	"tdameritrade-alerter/config"
)

const (
	baseUrl   = "https://api.tdameritrade.com/v1"
	chainsApi = "/marketdata/chains"
)

func main() {

	// executable is the first arg
	if len(os.Args) == 1 {
		log.Fatal("Please specify the path to the app config file")
	}

	cfg, err := LoadConfig(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to load the config " + err.Error())
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
		log.Fatal("Failed to invoke the API", err)
	}

	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to read response", err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("API call didn't succeed: " + string(respBytes))
	}
	chains := chain.Chains{}

	err = json.Unmarshal(respBytes, &chains)
	if err != nil {
		log.Fatal("Failed to parse the json response", err)
	}

	stdOutProcessor := chain.StdOutProcessor{cfg}
	_ = stdOutProcessor.Analyze(&chains)

	if cfg.SlackWebhookUrl == "" {
		log.Println("Slack webhook URL not configured")
	} else {
		slackProcessor := chain.SlackProcessor{cfg}
		err := slackProcessor.Analyze(&chains)
		if err != nil {
			log.Fatalf("Failed to notify via Slack %v", err.Error())
		}
	}
}

func LoadConfig(path string) (c config.Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	//viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&c)

	return
}
