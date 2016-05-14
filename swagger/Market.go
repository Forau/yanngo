package swagger

import ()

type Market struct {
	MarketId int64  `json:"market_id,omitempty"`
	Country  string `json:"country,omitempty"`
	Name     string `json:"name,omitempty"`
}
