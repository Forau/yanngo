package feed

import (
	"bytes"
	"encoding/json"
	"github.com/Forau/yanngo/feed/feedmodel"
	"github.com/Forau/yanngo/remote"
	"log"

	"sort"
)

// Convenience function to recive feed messages
type FeedClient func(msg *feedmodel.FeedMsg)

// Implement SubHandler
func (fc FeedClient) Handle(topic string, data []byte) (err error) {
	der := json.NewDecoder(bytes.NewReader(data))
	der.UseNumber()
	var msg feedmodel.FeedMsg
	if err = der.Decode(&msg); err == nil {
		fc(&msg)
	} else {
		log.Printf("[%s] WARNING: %v: (%s)", topic, err, string(data))
	}
	return
}

// Add the function to bind, for already existing references
func (fc FeedClient) Bind(ps remote.PubSub, feedUri string) error {
	return ps.Sub(feedUri, fc)
}

// And a new func to bind the traditional way
func BindFeedClient(ps remote.PubSub, feedUri string, fc FeedClient) error {
	return ps.Sub(feedUri, fc)
}

// FeedSorter will try and get out of order messages back in order.
// This can occure if you send the feed to a event queue.
// Requires that SeqId is set.
// Notice, we are not thread safe. If needed, add a FeedClient in between that only uses one go rutine to send.
// To flush cached data, send a nil.  (For scripts that backtests, to make sure all is delivered)
func FeedSorter(fc FeedClient) (ret FeedClient) {
	var waiting feedmodel.FeedMsgSeqSorter
	var lastSentSeq int64

	ret = func(msg *feedmodel.FeedMsg) {
		if msg == nil {
			// TODO: Dump all cached.
			for _, m := range waiting {
				fc(&m)
			}
			waiting = waiting[:0] // Clear, just to be sure we dont resend in canse of another nil
		} else if msg.SeqId <= 0 {
			// No seqid, so we exit here....
			fc(msg) // Just send it in the order we got it....
		} else {
			// If we are next element, or we are out of order with more then 1000, then we send.  (Diff over 1000 indicate a seq-reset)
			/// log.Printf("Seq %d, last %d. Queue: %d  [%d\n", msg.SeqId, lastSentSeq, len(waiting), msg.SeqId - lastSentSeq)
			if msg.SeqId == lastSentSeq+1 || msg.SeqId > lastSentSeq+1000 || msg.SeqId < lastSentSeq-1000 {
				lastSentSeq = msg.SeqId
				// log.Printf("Sending %d:\n",msg.SeqId)
				fc(msg)
				if len(waiting) > 0 {
					// Assume already sorted.
					*msg, waiting = waiting[0], waiting[1:]
					ret(msg) // And send the queued one.
				}
			} else {
				waiting = append(waiting, *msg)
				sort.Sort(waiting) // Expencive, but we shouldnt have too many. Out of order should be rare.
			}
		}
	}
	return
}

type timedFeedMsg struct {
	*feedmodel.FeedMsg
	timestamp  int64
	tiebreaker int64 // Mainly for price, where many trades can come clustered on sime tick.
}

type timeStruct struct {
	Tick_timestamp  int64
	Trade_timestamp int64
	Turnover_volume int64 // For now, use as tie-breaker for clustered prices.
}

func (ts *timeStruct) timestamp(defaultTs int64) int64 {
	if ts.Tick_timestamp > 0 {
		return ts.Tick_timestamp
	} else if ts.Trade_timestamp > 0 {
		return ts.Trade_timestamp
	}
	return defaultTs
}

type timedFeedMsgHolder []timedFeedMsg

func (tfmh timedFeedMsgHolder) Len() int      { return len(tfmh) }
func (tfmh timedFeedMsgHolder) Swap(i, j int) { tfmh[i], tfmh[j] = tfmh[j], tfmh[i] }
func (tfmh timedFeedMsgHolder) Less(i, j int) bool {
	return (tfmh[i].timestamp < tfmh[j].timestamp) ||
		(tfmh[i].timestamp == tfmh[j].timestamp && tfmh[i].tiebreaker < tfmh[j].tiebreaker)
}

// FeedSorterDelayedIfNotSeqId: A MUCH slower version of FeedSorter. Will delay messages by X millis,
// and then sort and deligate.
// Use mainly for historicaol data where SeqId might not exist, or for some reason the source might be scrambled.
// For a realtime experience, use normal FeedSorter if your messages are out of order.
func FeedSorterDelayedIfNotSeqId(delayMillis int64, fc FeedClient) (ret FeedClient) {
	seqIdSorter := FeedSorter(fc)

	var waiting timedFeedMsgHolder
	var lastestTimestamp int64

	ret = func(msg *feedmodel.FeedMsg) {
		if msg == nil {
			// TODO: Dump all cached.
			seqIdSorter(nil)
			for _, m := range waiting {
				fc(m.FeedMsg)
			}
			waiting = waiting[:0]
		} else if msg.SeqId > 0 {
			seqIdSorter(msg)
		} else {
			var timestruct timeStruct
			msg.DecodeData(&timestruct)
			ts := timestruct.timestamp(lastestTimestamp)

			if ts < lastestTimestamp-delayMillis {
				fc(msg) // Too old, send now...
			} else {
				waiting = append(waiting, timedFeedMsg{msg, ts, timestruct.Turnover_volume})
				if ts > lastestTimestamp {
					lastestTimestamp = ts
					sort.Sort(waiting)
					for idx, m := range waiting {
						if m.timestamp < lastestTimestamp-delayMillis {
							fc(m.FeedMsg)
						} else {
							waiting = waiting[idx:]
							return
						}
					}
					waiting = waiting[:0] // Reset, since we sent all
				}
			}
		}
	}
	return
}
