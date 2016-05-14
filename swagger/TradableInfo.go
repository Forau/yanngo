package swagger

import ()

type TradableInfo struct {
	MarketId   int64         `json:"market_id,omitempty"`
	Identifier string        `json:"identifier,omitempty"`
	Iceberg    bool          `json:"iceberg,omitempty"`
	Calendar   []CalendarDay `json:"calendar,omitempty"`
	OrderTypes []OrderType   `json:"order_types,omitempty"`
}
