package swagger

import ()

type PublicTrades struct {
	MarketId   int64         `json:"market_id,omitempty"`
	Identifier string        `json:"identifier,omitempty"`
	Trades     []PublicTrade `json:"trades,omitempty"`
}
