package swagger

type Tradable struct {
	MarketId     int64   `json:"market_id,omitempty"`
	Identifier   string  `json:"identifier,omitempty"`
	TickSizeId   int64   `json:"tick_size_id,omitempty"`
	LotSize      float64 `json:"lot_size,omitempty"`
	DisplayOrder int64   `json:"display_order,omitempty"`
}
