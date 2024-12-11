package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Config struct {
	Redis struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`
	Binance struct {
		BaseWSURL string   `json:"base_ws_url"`
		Symbols   []string `json:"symbols"`
	} `json:"binance"`
	DataRetention struct {
		KlineMaxItems    int `json:"kline_max_items"`
		TradesMaxItems   int `json:"trades_max_items"`
		OrderbookMaxSize int `json:"orderbook_max_size"`
	} `json:"data_retention"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %v", err)
	}

	log.Printf("Raw config loaded: BaseWSURL=%s", config.Binance.BaseWSURL)

	// Set defaults if not specified
	if config.Binance.BaseWSURL == "" {
		log.Printf("Setting default BaseWSURL")
		config.Binance.BaseWSURL = "wss://fstream.binance.com"
	}
	if config.DataRetention.KlineMaxItems == 0 {
		config.DataRetention.KlineMaxItems = 1000
	}
	if config.DataRetention.TradesMaxItems == 0 {
		config.DataRetention.TradesMaxItems = 5000
	}
	if config.DataRetention.OrderbookMaxSize == 0 {
		config.DataRetention.OrderbookMaxSize = 1000
	}

	return &config, nil
} 