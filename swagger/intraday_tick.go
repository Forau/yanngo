package swagger

type IntradayTick struct {
	Timestamp  int64   `json:"timestamp,omitempty"`
	Last       float64 `json:"last,omitempty"`
	Low        float64 `json:"low,omitempty"`
	High       float64 `json:"high,omitempty"`
	Volume     int64   `json:"volume,omitempty"`
	NoOfTrades int64   `json:"no_of_trades,omitempty"`
}
