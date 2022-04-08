package config

import (
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

// Config represents the config file structure
type Config struct {
	ImmoTrakt struct {
		Frequency             string `default:"1m" yaml:"frequency" envconfig:"IMMOTRAKT_FREQUENCY"`
		IncludeExistingOffers bool   `default:"false" yaml:"include_existing_offers" envconfig:"IMMOTRAKT_INCLUDE_EXISTING"`
	} `yaml:"immo_trakt"`
	Telegram struct {
		Token  string `yaml:"token" envconfig:"IMMOTRAKT_TELEGRAM_TOKEN"`
		ChatID string `yaml:"chat_id" envconfig:"IMMOTRAKT_TELEGRAM_CHAT_ID"`
	} `yaml:"telegram"`
	ImmobilienScout struct {
		Search        string `yaml:"search" envconfig:"IMMOTRAKT_SEARCH"`
		ExcludeWBS    bool   `default:"false" yaml:"exclude_wbs" envconfig:"IMMOTRAKT_EXCLUDE_WBS"`
		ExcludeTausch bool   `default:"false" yaml:"exclude_tausch" envconfig:"IMMOTRAKT_EXCLUDE_TAUSCH"`
		ExcludeSenior bool   `default:"false" yaml:"exclude_senior" envconfig:"IMMOTRAKT_EXCLUDE_SENIOR"`
	} `yaml:"immobilien_scout"`
}

// New creates config with values from both config file and environment
func New() (*Config, error) {
	var cfg Config

	err := readFile(&cfg)
	if err != nil {
		return nil, err
	}

	err = readEnv(&cfg)

	return &cfg, err
}

func readFile(config *Config) error {
	f, err := os.Open("config.yml")
	if err != nil {
		log.Println("config.yml is not found, as a backup we will try to load the values from environment variables.")
		return nil
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(config)

	return err
}

func readEnv(config *Config) error {
	err := envconfig.Process("", config)

	return err
}
