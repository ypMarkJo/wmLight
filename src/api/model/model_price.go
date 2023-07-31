package model

import "time"

type Price struct {
	Id int `json:"id"`

	Symbol string `json:"symbol" `

	Price float64 `json:"price"`

	// chainlink or bitfinex
	Source string `json:"source"`

	// timeStamp or blockNumber
	TimeStamp time.Time `json:"timestamp"`
}

func (p *Price) TableName() string {
	return "latest_price"
}
