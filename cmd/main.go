package main

import (
	"fmt"
	"os"
	"time"

	"github.com/DarkPhoenix42/p-torrent/pkg/torrent"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel string `yaml:"log_level"`
	LogFile  string `yaml:"log_file"`
	MaxPeers int    `yaml:"max_peers"`
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	config, err := loadConfig("../config.yaml")

	if err != nil {
		fmt.Println("Config file not found!")
		return
	}

	log_level, err := zerolog.ParseLevel(config.LogLevel)

	if err != nil {
		fmt.Println("Failed to parse log level! Defaulting to INFO level.")
		log_level = zerolog.InfoLevel
	}

	logger := zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.DateTime},
	).Level(log_level).With().Timestamp().Caller().Logger()

	logger.Info().Msg("Initialized logger!")

	file_name := os.Args[1]
	_, err = torrent.NewTorrent(file_name)
	if err != nil {
		logger.Error().Msg("Failed to create a torrent!")
	}

	logger.Info().Msg("Successful!")
}
