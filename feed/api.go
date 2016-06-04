package feed

import (
	"encoding/json"
	"fmt"
	"github.com/Forau/yanngo/api"
	"log"
	"time"
)

// Used when sending feed commands
type feedCmd struct {
	Cmd  string      `json:"cmd"`
	Args interface{} `json:"args"`
}

type feedCmdArgs struct {
	T string `json:"t"`
	I string `json:"i"`
	M int64  `json:"m"`
}

// Arguments for subscribing to indicator updates
type indicatorArgs struct {
	T string `json:"t"`
	I string `json:"i"`
	M string `json:"m"`
}

// Arguments for subscribing to news updates
type newsArgs struct {
	T     string `json:"t"`
	S     int64  `json:"s"`
	Delay bool   `json:"delay,omitempty"`
}

type FeedMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (fm *FeedMessage) String() string {
	b, _ := json.Marshal(fm)
	return string(b)
}

// This object should connect to the feeds, keep them alive, and keep state
type FeedDaemon struct {
	api     *api.ApiClient
	private *baseFeed
	public  *baseFeed
}

// This should be a signleton, so might change initiation later.
func NewFeedDaemon(api *api.ApiClient) (*FeedDaemon, error) {
	fd := &FeedDaemon{api: api}
	go fd.privatFeedLoop()
	go fd.publicFeedLoop()
	return fd, nil
}

func (fd *FeedDaemon) privatFeedLoop() {
	// TODO: Handle death
	for {
		session, err := fd.api.Session() // TODO: Error check

		fd.private, err = newBaseFeed(&session, session.PrivateFeed)
		if err != nil {
			log.Printf("Error creating feed, sleeping: ", err)
			time.Sleep(5000 * time.Millisecond)
		} else {
			fd.feedLoop(fd.private)
		}
	}
}

// TODO: Join with private
func (fd *FeedDaemon) publicFeedLoop() {
	// TODO: Handle death
	for {
		session, err := fd.api.Session() // TODO: Error check
		fd.public, err = newBaseFeed(&session, session.PublicFeed)
		if err != nil {
			log.Printf("Error creating feed, sleeping: ", err)
			time.Sleep(5000 * time.Millisecond)
		} else {
			fd.feedLoop(fd.public)
		}
	}
}

func (fd *FeedDaemon) feedLoop(feed *baseFeed) {
	for {
		msg := &FeedMessage{}
		if err := feed.decoder.Decode(msg); err != nil {
			log.Printf("Got error: %+v", err)
			return
		} else {
			if msg.Type != "heartbeat" {
				log.Printf("Got message: %+v", msg)
			}
		}
	}
}

func makeArgs(typ, ident, market string) (ret interface{}, err error) {
	switch typ {
	case "price", "depth", "trade", "trading_status":
		args := &feedCmdArgs{T: typ, I: ident}
		_, err = fmt.Sscan(market, &args.M)
		ret = args
	case "indicator":
		ret = &indicatorArgs{T: typ, I: ident, M: market}
	case "news":
		ret = &newsArgs{T: typ} // TODO: Separate news?
	}
	return
}

func (fd *FeedDaemon) Subscribe(typ, ident, market string) error {
	args, e := makeArgs(typ, ident, market)
	if e != nil {
		return e
	}
	cmd := &feedCmd{
		Cmd:  "subscribe",
		Args: args,
	}
	return fd.public.Write(cmd)
}
