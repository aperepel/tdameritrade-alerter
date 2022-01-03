package util

import (
	"github.com/spf13/viper"
	"os"
	"tdameritrade-alerter/config"
)

// Will return false when running in a k8s container
func IsStandalone() bool {
	// will be mounted via k8s downward api
	_, err := os.Stat("/etc/podinfo")
	if os.IsNotExist(err) {
		return true // standalone
	} else {
		return false // k8s
	}
}

func LoadConfig(path string) (c config.Config, err error) {
	// override couple values from env if configured
	viper.AutomaticEnv()
	_ = viper.BindEnv("ApiKey", "API_KEY")
	_ = viper.BindEnv("SlackWebhookUrl", "SLACK_WEBHOOK_URL")
	viper.AddConfigPath(path)
	viper.SetConfigName("alert")
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	// merge in the secrets config
	viper.SetConfigName(".secrets")
	err = viper.MergeInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&c)

	return
}
