package feedmodel

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type FeedType uint64

const (
	PrivateFeedType FeedType = iota + 1
	PublicFeedType
)

// Used when sending feed commands
type FeedCmd struct {
	Cmd  string      `json:"cmd"`
	Args interface{} `json:"args"`
}

type FeedMsg struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func NewFeedMsg(data []byte) (ret *FeedMsg, err error) {
	ret = &FeedMsg{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	err = dec.Decode(ret)
	return
}

func NewFeedMsgFromObject(typ string, data interface{}) (ret *FeedMsg, err error) {
	ret = &FeedMsg{Type: typ}
	ret.Data, err = json.Marshal(data)
	return
}

func (fm *FeedMsg) String() string {
	return fmt.Sprintf("FeedMsg[%s]: %s", fm.Type, string(fm.Data))
}

func (fm *FeedMsg) DecodeData(ret interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(fm.Data))
	dec.UseNumber()
	return dec.Decode(ret)
}

// TODO: See if we want to actually encode, or just fake
func (fm *FeedMsg) Encode() (ret []byte) {
	//	return []byte(fmt.Sprintf(`{"type":"%s","data":"%s"}`, fm.Type, string(fm.Data)))
	ret, _ = json.Marshal(fm)
	return
}

// ---------- Some type specific messages

// Struct to reprecent a price update
type FeedPriceData struct {
	Identifier      string  `json:"i,omitempty"`
	Market          int64   `json:"m,omitempty"`
	Ask             float64 `json:"ask,omitempty"`
	Ask_volume      float64 `json:"ask_volume,omitempty"`
	Bid             float64 `json:"bid,omitempty"`
	Bid_volume      float64 `json:"bid_volume,omitempty"`
	Open            float64 `json:"open,omitempty"`
	High            float64 `json:"high,omitempty"`
	Low             float64 `json:"low,omitempty"`
	Last            float64 `json:"last,omitempty"`
	Last_volume     int64   `json:"last_volume,omitempty"`
	Close           float64 `json:"close,omitempty"`
	EP              float64 `json:"ep,omitempty"`
	Imbalance       int64   `json:"imbalance,omitempty"`
	Paired          int64   `json:"paired,omitempty"`
	Turnover        float64 `json:"turnover,omitempty"`
	Turnover_volume int64   `json:"turnover_volume,omitempty"`
	Vwap            float64 `json:"vwap,omitempty"`
	Tick_timestamp  int64   `json:"tick_timestamp,omitempty"`
	Trade_timestamp int64   `json:"trade_timestamp,omitempty"`

	prev *FeedPriceData // Optional use. Can be used to 'link' prices together, and from that determine if this is a trade or not
}

// Usage: fpd = fpd.Append(newFpd), like on slices
func (fpd *FeedPriceData) Append(fp *FeedPriceData) (ret *FeedPriceData) {
	// This trade hapend before.  (We might need to check for day too). We _can_ be nil
	if fpd != nil && fpd.Turnover_volume > fp.Turnover_volume {
		if fpd.prev != nil {
			fpd.prev = fpd.prev.Append(fp) // Move it back
		} else {
			fpd.prev = fp
		}
		ret = fpd
	} else {
		fp.prev, ret = fpd, fp
	}
	return
}

func (fpd *FeedPriceData) ToSlice() []*FeedPriceData {
	return fpd.ToLimitedSlice(func(in *FeedPriceData) (include, brk bool) {
		return true, fpd.prev == nil
	})
}

func (fpd *FeedPriceData) ToLimitedSlice(fbm FilterBreakMatcher) (ret []*FeedPriceData) {
	inc, brk := fbm(fpd)

	if !brk && fpd.prev != nil {
		ret = fpd.prev.ToLimitedSlice(fbm)
	}

	if inc {
		ret = append(ret, fpd)
	}
	return
}

func (fpd *FeedPriceData) Each(fn func(in *FeedPriceData)) {
	for _, fp := range fpd.ToSlice() {
		fn(fp)
	}
}

func (fpd *FeedPriceData) IsTrade() bool {
	if fpd.Tick_timestamp == fpd.Trade_timestamp {
		return fpd.prev == nil || (fpd.Turnover_volume != fpd.prev.Turnover_volume) // Only check if we are linked
	}
	return false
}

// Filters

// Filter type, return true on include if matches, and true on brk if we need to break
type FilterBreakMatcher func(in *FeedPriceData) (include, brk bool)

func (fbm FilterBreakMatcher) Filter(fn FilterBreakMatcher) FilterBreakMatcher {
	return func(in *FeedPriceData) (include, brk bool) {
		inc, brk := fbm(in)
		if inc {
			i, b := fn(in)
			return i, b || brk
		}
		return inc, brk
	}
}

func FilterOnlyLastPerTickTimestamp() FilterBreakMatcher {
	var lastTick *FeedPriceData
	return func(in *FeedPriceData) (include, brk bool) {
		if lastTick == nil || lastTick.Tick_timestamp != in.Tick_timestamp {
			include = true
			lastTick = in
		}
		return
	}
}

func FilterTrades() FilterBreakMatcher {
	return func(in *FeedPriceData) (include, brk bool) {
		return in.IsTrade(), false
	}
}

func FilterCount(num int64) FilterBreakMatcher {
	var count int64
	return func(in *FeedPriceData) (include, brk bool) {
		count++
		if count >= num {
			return true, true
		}
		return true, false
	}
}
