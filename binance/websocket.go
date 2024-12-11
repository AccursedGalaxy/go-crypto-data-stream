package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type WebsocketClient struct {
	baseURL     string
	symbols     []string
	conn        *websocket.Conn
	mu          sync.Mutex
	handlers    map[string]func([]byte) error
}

func NewWebsocketClient(baseURL string, symbols []string) *WebsocketClient {
	return &WebsocketClient{
		baseURL:  baseURL,
		symbols:  symbols,
		handlers: make(map[string]func([]byte) error),
	}
}

func (w *WebsocketClient) Connect(ctx context.Context) error {
	streams := w.buildStreamNames()
	
	url := fmt.Sprintf("%s/stream?streams=%s", w.baseURL, strings.Join(streams, "/"))

	log.Printf("Connecting to WebSocket URL: %s", url)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Binance websocket: %v", err)
	}

	w.mu.Lock()
	w.conn = conn
	w.mu.Unlock()

	return nil
}

func (w *WebsocketClient) RegisterHandler(streamType string, handler func([]byte) error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers[streamType] = handler
}

func (w *WebsocketClient) Listen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, message, err := w.conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading websocket message: %v", err)
				return err
			}

			// Log the raw message for debugging
			log.Printf("Raw message received: %s", string(message))

			// Try to unmarshal as a stream message
			var streamMsg struct {
				Stream string          `json:"stream"`
				Data   json.RawMessage `json:"data"`
			}

			if err := json.Unmarshal(message, &streamMsg); err != nil {
				// Try to unmarshal as a direct message
				var directMsg map[string]interface{}
				if err := json.Unmarshal(message, &directMsg); err != nil {
					log.Printf("Error unmarshaling message: %v", err)
					continue
				}

				// Handle direct message format
				if eventType, ok := directMsg["e"].(string); ok {
					if handler, ok := w.handlers[eventType]; ok {
						if err := handler(message); err != nil {
							log.Printf("Error handling direct message: %v", err)
						}
					} else {
						log.Printf("No handler registered for event type: %s", eventType)
					}
				}
				continue
			}

			// Handle stream message format
			parts := strings.Split(streamMsg.Stream, "@")
			if len(parts) < 2 {
				log.Printf("Invalid stream format: %s", streamMsg.Stream)
				continue
			}

			streamType := parts[1]
			if strings.Contains(streamType, "@") {
				streamType = strings.Split(streamType, "@")[0]
			}

			if handler, ok := w.handlers[streamType]; ok {
				if err := handler(streamMsg.Data); err != nil {
					log.Printf("Error handling message: %v", err)
				}
			} else {
				log.Printf("No handler registered for stream type: %s", streamType)
			}
		}
	}
}

func (w *WebsocketClient) buildStreamNames() []string {
	var streams []string
	for _, symbol := range w.symbols {
		symbol = strings.ToLower(symbol)
		// Kline stream for 1m interval
		streams = append(streams, fmt.Sprintf("%s@kline_1m", symbol))
		// Aggregate trade streams (using aggTrade instead of trade for futures)
		streams = append(streams, fmt.Sprintf("%s@aggTrade", symbol))
		// Book ticker streams
		streams = append(streams, fmt.Sprintf("%s@bookTicker", symbol))
		// Partial depth streams
		streams = append(streams, fmt.Sprintf("%s@depth20@100ms", symbol))
	}
	return streams
}

func (w *WebsocketClient) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
} 