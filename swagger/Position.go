package swagger

import ()

type Position struct {
	Accno          int32      `json:"accno,omitempty"`
	Instrument     Instrument `json:"instrument,omitempty"`
	Qty            float32    `json:"qty,omitempty"`
	PawnPercent    int32      `json:"pawn_percent,omitempty"`
	MarketValueAcc Amount     `json:"market_value_acc,omitempty"`
	MarketValue    Amount     `json:"market_value,omitempty"`
	AcqPrice       Amount     `json:"acq_price,omitempty"`
	AcqPriceAcc    Amount     `json:"acq_price_acc,omitempty"`
	MorningPrice   Amount     `json:"morning_price,omitempty"`
}
