package models

import "time"

// Kline represents a candlestick
type Kline struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Kline     struct {
		StartTime    int64  `json:"t"`
		CloseTime    int64  `json:"T"`
		Symbol       string `json:"s"`
		Interval     string `json:"i"`
		FirstTradeID int64  `json:"f"`
		LastTradeID  int64  `json:"L"`
		OpenPrice    string `json:"o"`
		ClosePrice   string `json:"c"`
		HighPrice    string `json:"h"`
		LowPrice     string `json:"l"`
		Volume       string `json:"v"`
		TradeCount   int    `json:"n"`
		IsClosed     bool   `json:"x"`
		QuoteVolume  string `json:"q"`
		TakerVolume  string `json:"V"`
		TakerQuoteVolume string `json:"Q"`
	} `json:"k"`
}

// Trade represents a single trade
type Trade struct {
	EventType string    `json:"e"`
	EventTime time.Time `json:"E"`
	Symbol    string    `json:"s"`
	TradeID   int64     `json:"a"`
	Price     string    `json:"p"`
	Quantity  string    `json:"q"`
	FirstID   int64     `json:"f"`
	LastID    int64     `json:"l"`
	TradeTime time.Time `json:"T"`
	IsMaker   bool      `json:"m"`
}

// OrderBookLevel represents a single price level in the order book
type OrderBookLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// OrderBook represents the full order book
type OrderBook struct {
	Symbol    string           `json:"symbol"`
	LastUpdateID int64        `json:"lastUpdateId"`
	Bids      []OrderBookLevel `json:"bids"`
	Asks      []OrderBookLevel `json:"asks"`
	Timestamp time.Time        `json:"timestamp"`
}

// StreamMessage represents a generic websocket message
type StreamMessage struct {
	Stream string      `json:"stream"`
	Data   interface{} `json:"data"`
}

// BookTicker struct for futures book ticker
type BookTicker struct {
	EventType     string `json:"e"`
	UpdateID      int64  `json:"u"`
	EventTime     int64  `json:"E"`
	TransactionTime int64 `json:"T"`
	Symbol        string `json:"s"`
	BestBidPrice  string `json:"b"`
	BestBidQty    string `json:"B"`
	BestAskPrice  string `json:"a"`
	BestAskQty    string `json:"A"`
} 