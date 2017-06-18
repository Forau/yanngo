// Misc utils. First up is a small wrapper for TicksizeTables, where simple calculations can be made.
package nnutils

import (
	"github.com/Forau/yanngo/swagger"

	"fmt"
	"math"
	"sort"
)

var (
	// This is the most common tick size. Hardcoded here to avoid hitting the db or api for a basic configuration.
	DEFAULT_TICK_TABLE = swagger.TicksizeTable{
		TickSizeId: 14369,
		Ticks: []swagger.TicksizeInterval{
			{Decimals: 3, FromPrice: 0, ToPrice: 0.499, Tick: 0.001},
			{Decimals: 3, FromPrice: 0.5, ToPrice: 0.995, Tick: 0.005},
			{Decimals: 2, FromPrice: 1, ToPrice: 4.99, Tick: 0.01},
			{Decimals: 2, FromPrice: 15, ToPrice: 49.9, Tick: 0.1},
			{Decimals: 2, FromPrice: 150, ToPrice: 499.5, Tick: 0.5},
			{Decimals: 2, FromPrice: 5, ToPrice: 14.95, Tick: 0.05},
			{Decimals: 2, FromPrice: 50, ToPrice: 149.75, Tick: 0.25},
			{Decimals: 0, FromPrice: 500, ToPrice: 4999, Tick: 1},
			{Decimals: 0, FromPrice: 5000, ToPrice: 499995, Tick: 5},
		}}
)

// New name, so we dont need to expose swagger package. Assume sorted
type TickTableUtil []swagger.TicksizeInterval

func (ttu TickTableUtil) Len() int           { return len(ttu) }
func (ttu TickTableUtil) Swap(i, j int)      { ttu[i], ttu[j] = ttu[j], ttu[i] }
func (ttu TickTableUtil) Less(i, j int) bool { return ttu[i].FromPrice < ttu[j].FromPrice }

func NewDefaultTickTableUtil() (ret TickTableUtil) {
	ret = TickTableUtil(DEFAULT_TICK_TABLE.Ticks)
	sort.Sort(ret)
	return
}

func (ttu TickTableUtil) AsTick(data float64) (newVal, tick, from, to float64, decimals int64) {
	for i := len(ttu) - 1; i >= 0; i-- {
		if data >= ttu[i].FromPrice {
			tick = ttu[i].Tick
			from = ttu[i].FromPrice
			to = ttu[i].ToPrice
			decimals = ttu[i].Decimals

			newVal = math.Trunc(data/tick) * tick
			return
		}
	}
	return
}

func (ttu TickTableUtil) AddTicks(data float64, ticks int64) float64 {
	newVal, tick, from, to, _ := ttu.AsTick(data)

	for ticks != 0 {
		if ticks > 0 {
			ticks--
			newVal += tick
			if newVal > to {
				newVal, tick, from, to, _ = ttu.AsTick(newVal)
			}
		} else {
			ticks++
			newVal -= tick
			if newVal < from {
				newVal, tick, from, to, _ = ttu.AsTick(newVal)
			}
		}
	}

	return newVal
}

func (ttu TickTableUtil) ToString(data float64) string {
	newVal, _, _, _, dec := ttu.AsTick(data)
	return fmt.Sprintf(fmt.Sprintf("%%.%df", dec), newVal)
}

func (ttu TickTableUtil) TicksBetween(data1, data2 float64) (ret int64) {
	dir := int64(1)
	// Swap dir
	if data1 > data2 {
		data1, data2, dir = data2, data1, -dir
	}

	newVal, tick, _, to, _ := ttu.AsTick(data1)
	for {
		if data2 < to || newVal >= data2 {
			diffTicks := (data2 - newVal) / tick
			return dir * (ret + int64(diffTicks))
		} else {
			diffTicks := ((to - newVal) / tick) + 1
			newVal = to + tick*1
			ret += int64(diffTicks)
			lastTo := to
			newVal, tick, _, to, _ = ttu.AsTick(newVal)
			if to == lastTo {
				to = data2 + tick // We are on the highest tick level, so raise to so it fits
			}
		}
	}
}
