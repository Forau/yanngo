package swagger

type TicksizeInterval struct {
	Decimals  int64   `json:"decimals,omitempty"`
	FromPrice float64 `json:"from_price,omitempty"`
	ToPrice   float64 `json:"to_price,omitempty"`
	Tick      float64 `json:"tick,omitempty"`
}
