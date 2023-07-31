package model

type TickerResponse struct {
	LastPrice string `json:"last_price"`
	Timestamp string `json:"timestamp,omitempty"`
}
