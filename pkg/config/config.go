package config

import "github.com/spf13/viper"

type KafkaConfig struct {
	Brokers []string     `mapstructure:"brokers"`
	Reader  ReaderConfig `mapstructure:"reader"`
	Writer  WriterConfig `mapstructure:"writer"`
}

type ReaderConfig struct {
	Topic   string `mapstructure:"topic"`
	GroupID string `mapstructure:"group_id"`
}

type WriterConfig struct {
	Topic string `mapstructure:"topic"`
}

type Config struct {
	Kafka KafkaConfig `mapstructure:"kafka"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
