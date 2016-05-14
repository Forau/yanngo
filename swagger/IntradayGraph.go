package swagger

import ()

type IntradayGraph struct {
	MarketId   int64          `json:"market_id,omitempty"`
	Identifier string         `json:"identifier,omitempty"`
	Ticks      []IntradayTick `json:"ticks,omitempty"`
}
