package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/accursedgalaxy/marketmaker/go-crypto-data-stream/binance"
	"github.com/accursedgalaxy/marketmaker/go-crypto-data-stream/config"
	"github.com/accursedgalaxy/marketmaker/go-crypto-data-stream/models"
	"github.com/accursedgalaxy/marketmaker/go-crypto-data-stream/storage"
)

func main() {
	// Load configuration
	log.Printf("Attempting to load config from: %s", "config.json")
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Add detailed debug logging
	log.Printf("Config file loaded successfully")
	log.Printf("Redis config: host=%s, port=%d", cfg.Redis.Host, cfg.Redis.Port)
	log.Printf("Binance config: url=%s", cfg.Binance.BaseWSURL)
	log.Printf("Binance symbols: %v", cfg.Binance.Symbols)

	// Create a context that will be canceled on program termination
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Redis client
	redisClient, err := storage.NewRedisClient(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize Binance websocket client
	wsClient := binance.NewWebsocketClient(cfg.Binance.BaseWSURL, cfg.Binance.Symbols)

	// Register handlers for different stream types
	wsClient.RegisterHandler("kline_1m", func(data []byte) error {
		var kline models.Kline
		if err := json.Unmarshal(data, &kline); err != nil {
			return err
		}
		return redisClient.StoreKline(ctx, kline.Symbol, "1m", kline)
	})

	wsClient.RegisterHandler("aggTrade", func(data []byte) error {
		var trade models.Trade
		if err := json.Unmarshal(data, &trade); err != nil {
			return err
		}
		return redisClient.StoreTrade(ctx, trade.Symbol, trade)
	})

	wsClient.RegisterHandler("bookTicker", func(data []byte) error {
		var bookTicker models.BookTicker
		if err := json.Unmarshal(data, &bookTicker); err != nil {
			return err
		}
		return redisClient.StoreBookTicker(ctx, bookTicker.Symbol, bookTicker)
	})

	wsClient.RegisterHandler("depth20", func(data []byte) error {
		var orderbook models.OrderBook
		if err := json.Unmarshal(data, &orderbook); err != nil {
			return err
		}
		return redisClient.StoreOrderBook(ctx, orderbook.Symbol, orderbook)
	})

	// Connect to Binance websocket
	if err := wsClient.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to Binance websocket: %v", err)
	}
	defer wsClient.Close()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Starting Binance data stream service...")

	// Start listening for websocket messages in a goroutine
	go func() {
		if err := wsClient.Listen(ctx); err != nil {
			log.Printf("Websocket listener error: %v", err)
			cancel()
		}
	}()

	// Wait for termination signal
	<-sigChan
	log.Println("Shutting down gracefully...")
} 