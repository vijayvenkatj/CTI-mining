package config

import (
	"time"

	"github.com/spf13/viper"
)

type ReaderConfig struct {
	Topic   string `mapstructure:"topic"`
	GroupID string `mapstructure:"group_id"`
}

type WriterConfig struct {
	Topic string `mapstructure:"topic"`
}

type UsecaseConfig struct {
	Reader *ReaderConfig `mapstructure:"reader"`
	Writer *WriterConfig `mapstructure:"writer"`
}

type KafkaConfig struct {
	Brokers       []string      `mapstructure:"brokers"`
	EdgeGenerator UsecaseConfig `mapstructure:"edge_generator"`
	Estimator     UsecaseConfig `mapstructure:"estimator"`
}

type OTXConfig struct {
	APIKey        string `mapstructure:"api_key"`
	BaseURL       string `mapstructure:"base_url"`
	ModifiedSince string `mapstructure:"modified_since"`

	InitialBackoff time.Duration `mapstructure:"initial_backoff"`
	MaxBackoff     time.Duration `mapstructure:"max_backoff"`
}

type Config struct {
	Kafka KafkaConfig `mapstructure:"kafka"`
	OTX   OTXConfig   `mapstructure:"otx"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	v.SetDefault("otx.base_url", "https://otx.alienvault.com/api/v1/pulses/subscribed")
	v.SetDefault("otx.modified_since", "2026-09-01T00:00:00Z")
	v.SetDefault("otx.initial_backoff", time.Second)
	v.SetDefault("otx.max_backoff", time.Hour)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
