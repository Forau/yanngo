package feedmodel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"log"
)

type FeedType uint64

const (
	PrivateFeedType FeedType = iota + 1
	PublicFeedType
)

type TradeSide int64

const (
	BID TradeSide = iota - 1
	Unknown
	ASK
)

func (ts TradeSide) String() string {
	switch ts {
	case BID:
		return "BID"
	case ASK:
		return "ASK"
	default:
		return "Unknown"
	}
}

// Used when sending feed commands
type FeedCmd struct {
	Cmd  string      `json:"cmd"`
	Args interface{} `json:"args"`
}

type FeedMsg struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`

	SeqId int64 `json:"seq,omitempty"` // Optional field that we can add local sequence to
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
	if fm == nil {
		return fmt.Errorf("FeedMsg was nil, but still tried to decode")
	}
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

type FeedMsgSeqSorter []FeedMsg

func (fmss FeedMsgSeqSorter) Len() int           { return len(fmss) }
func (fmss FeedMsgSeqSorter) Swap(i, j int)      { fmss[i], fmss[j] = fmss[j], fmss[i] }
func (fmss FeedMsgSeqSorter) Less(i, j int) bool { return fmss[i].SeqId < fmss[j].SeqId }

// ---------- Some type specific messages

type FeedPriceDataSorter []FeedPriceData

func (fpds FeedPriceDataSorter) Len() int           { return len(fpds) }
func (fpds FeedPriceDataSorter) Swap(i, j int)      { fpds[i], fpds[j] = fpds[j], fpds[i] }
func (fpds FeedPriceDataSorter) Less(i, j int) bool { return fpds[i].Less(&fpds[j]) }

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

func (fpd *FeedPriceData) Less(other *FeedPriceData) bool {
	return (fpd.Tick_timestamp < other.Tick_timestamp) ||
		(fpd.Tick_timestamp == other.Tick_timestamp && fpd.Turnover_volume < other.Turnover_volume)
}

// Usage: fpd = fpd.Append(newFpd), like on slices
func (fpd *FeedPriceData) Append(fp *FeedPriceData) (ret *FeedPriceData) {
	// This trade hapend before.  (We might need to check for day too). We _can_ be nil
	if fpd != nil && fp.Less(fpd) {
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

func (fpd *FeedPriceData) TradeSide() TradeSide {
	if fpd.IsTrade() && fpd.prev != nil {
		if fpd.prev.Ask == fpd.Last {
			return ASK
		} else if fpd.prev.Bid == fpd.Last {
			return BID
		}
	}
	return Unknown
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

type FeedTradeData struct {
	Identifier      string  `json:"i,omitempty"`
	Market          int64   `json:"m,omitempty"`
	Trade_timestamp int64   `json:"trade_timestamp,omitempty"`
	Trade_type      string  `json:"trade_type,omitempty"`
	Trade_id        string  `json:"trade_id,omitempty"`
	Price           float64 `json:"price,omitempty"`
	Volume          int64   `json:"volume,omitempty"`
	Broker_buying   string  `json:"broker_buying,omitempty"`
	Broker_selling  string  `json:"broker_selling,omitempty"`
}

// Like strconv.Atoi, but more hack 'n slash
func (ftd *FeedTradeData) TradeIdNum() (ret int64, ok bool) {
	ok = true // Default true
	for _, b := range []byte(ftd.Trade_id) {
		ret *= 10 // To shift up last one
		switch {
		case '0' <= b && b <= '9':
			ret += int64(b - '0')
		default:
			ok = false
			return
		}
	}
	return
}

type FeedTradeDataSequence []FeedTradeData

func (ftds FeedTradeDataSequence) Len() int      { return len(ftds) }
func (ftds FeedTradeDataSequence) Swap(i, j int) { ftds[i], ftds[j] = ftds[j], ftds[i] }
func (ftds FeedTradeDataSequence) Less(i, j int) bool {
	return ftds[i].Trade_timestamp < ftds[j].Trade_timestamp ||
		(ftds[i].Trade_timestamp == ftds[j].Trade_timestamp && ftds[i].Trade_id < ftds[j].Trade_id)
}

func (ftds FeedTradeDataSequence) Tail(num int) (ret FeedTradeDataSequence) {
	if l := len(ftds); num >= l {
		return ftds
	} else {
		return ftds[l-num:]
	}
}

type FeedIndicatorData struct {
	Identifier     string  `json:"i,omitempty"`
	Market         string  `json:"m,omitempty"`
	Tick_timestamp int64   `json:"tick_timestamp,omitempty"`
	Open           float64 `json:"open,omitempty"`
	High           float64 `json:"high,omitempty"`
	Low            float64 `json:"low,omitempty"`
	Last           float64 `json:"last,omitempty"`
	Close          float64 `json:"close,omitempty"`
}

type FeedIndicatorDataSequence []FeedIndicatorData

func (fids FeedIndicatorDataSequence) Len() int      { return len(fids) }
func (fids FeedIndicatorDataSequence) Swap(i, j int) { fids[i], fids[j] = fids[j], fids[i] }
func (fids FeedIndicatorDataSequence) Less(i, j int) bool {
	return fids[i].Tick_timestamp < fids[j].Tick_timestamp
}

func (fids FeedIndicatorDataSequence) GroupMillis(m int64) (ret []FeedIndicatorDataSequence) {
	var lastTime int64
	var idx int

	for _, fid := range fids {
		tid := fid.Tick_timestamp / m
		if tid != lastTime {
			ret = append(ret, FeedIndicatorDataSequence{fid})
			idx = len(ret) - 1
			lastTime = tid
		} else {
			// If first time is 0, we will hit nil
			ret[idx] = append(ret[idx], fid)
		}
	}
	return
}

func (fids FeedIndicatorDataSequence) Last() (ret *FeedIndicatorData) {
	if l := len(fids); l > 0 {
		ret = &(fids[l-1])
	}
	return
}

func (fids FeedIndicatorDataSequence) At(timestamp int64) (ret *FeedIndicatorData) {
	for i := len(fids) - 1; i >= 0; i-- {
		if fids[i].Tick_timestamp <= timestamp {
			ret = &(fids[i])
			return
		}
	}
	return
}
func (fids FeedIndicatorDataSequence) Sort() {
	sort.Sort(fids)
}

/// A struct like swagger.TradableId, since we dont want to import that package here.
type tradableId struct {
	identifier string
	market     int64
}

// TradeCatcher: Will try and extract trades from a price stream.
// Will also feed normal trades, so if you want to filter, the type on all generated is 'TradeCatcher'
// A feed listener that will forward trades, or fake trades depending on submit on price or trade
type TradeCatcher struct {
	lastPrice map[tradableId]*FeedPriceData

	tradeChan func(*FeedTradeData) // Since we only allow one to one, chan is not as useful
}

// Creates a TradeCatcher
func NewTradeCatcher(callback func(*FeedTradeData)) (ret *TradeCatcher) {
	ret = &TradeCatcher{lastPrice: make(map[tradableId]*FeedPriceData), tradeChan: callback}
	return
}

// Implement FeedClient.  If from async source, then wrap with feed.FeedSorter
func (tc *TradeCatcher) OnFeed(msg *FeedMsg) {
	if msg.Type == "price" {
		var price FeedPriceData
		if err := msg.DecodeData(&price); err == nil {
			tc.OnPrice(&price)
		} else {
			log.Printf("Could not decode price data: %+v", err)
		}
	} else if msg.Type == "trade" {
		var trade FeedTradeData
		if err := msg.DecodeData(&trade); err == nil {
			tc.OnTrade(&trade)
		} else {
			log.Printf("Could not decode trade data: %+v", err)
		}
	}
}

// Assume in order..
func (tc *TradeCatcher) OnPrice(price *FeedPriceData) {
	key := tradableId{price.Identifier, price.Market}
	if price.Trade_timestamp == price.Tick_timestamp {
		if lp, ok := tc.lastPrice[key]; ok {
			if lp.Turnover_volume < price.Turnover_volume {
				// TRADE!!!
				trade := &FeedTradeData{
					Identifier:      price.Identifier,
					Market:          price.Market,
					Trade_timestamp: price.Trade_timestamp,
					Trade_type:      "TradeCatcher",
					Trade_id:        fmt.Sprintf("%f - %f: %f -- %d", price.Bid, price.Ask, price.Vwap, price.Turnover_volume),
					Price:           price.Last,
					Volume:          price.Turnover_volume - lp.Turnover_volume, // Use diff, since we can resort, and last migh be wron
				}
				tc.OnTrade(trade)
			}
		}
	}
	tc.lastPrice[key] = price
}

func (tc *TradeCatcher) OnTrade(trade *FeedTradeData) {
	tc.tradeChan(trade)
}
