package swagger

type IntradayTick struct {
	Timestamp  int32   `json:"timestamp,omitempty"`
	Last       float64 `json:"last,omitempty"`
	Low        float64 `json:"low,omitempty"`
	High       float64 `json:"high,omitempty"`
	Volume     int32   `json:"volume,omitempty"`
	NoOfTrades int32   `json:"no_of_trades,omitempty"`
}
