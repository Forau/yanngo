package swagger

import ()

type TradableId struct {
	Identifier string `json:"identifier,omitempty"`
	MarketId   int64  `json:"market_id,omitempty"`
}
