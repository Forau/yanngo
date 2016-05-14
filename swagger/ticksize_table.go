package swagger

import ()

type TicksizeTable struct {
	TickSizeId int64              `json:"tick_size_id,omitempty"`
	Ticks      []TicksizeInterval `json:"ticks,omitempty"`
}
