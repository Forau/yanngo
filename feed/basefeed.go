package feed

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	//	"github.com/Forau/yanngo/swagger"
	"bytes"
	"io"
	"log"
	"time"
)

// Set this static variable if you need special TLS config.  Used in tests for example.
var DefaultTLS *tls.Config

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
}

type CmdWriter func(cmd *FeedCmd) error
type FeedType uint64

const (
	PrivateFeedType FeedType = iota + 1
	PublicFeedType
)

type Callback interface {
	OnConnect(w CmdWriter, ft FeedType)
	OnMessage(msg *FeedMsg, ft FeedType)
	OnError(err error, ft FeedType)
}

type SessionProvider func() (key, url string, err error)

// Arguments for sending the login command
type loginArgs struct {
	SessionKey string `json:"session_key"`
}

func makeDelayRetry(msg string) func() {
	resetTimer := time.Now().Add(time.Minute)
	counter := 0
	return func() {
		if time.Now().After(resetTimer) {
			counter = 0
		}
		resetTimer = time.Now().Add(time.Minute) // Reset in one minute, if no other call
		counter++
		var delay time.Duration
		switch counter {
		case 1:
			delay = 0
		case 2:
			delay = 5000
		default:
			delay = 30000
		}
		log.Printf(msg, delay)
		time.Sleep(delay * time.Millisecond)
	}
}

type baseFeed struct {
	conn    io.ReadWriteCloser
	encoder *json.Encoder
	decoder *json.Decoder

	feedType FeedType
	callback Callback
	quit     chan interface{}
}

// For debugging
type ConnWrap struct {
	conn io.ReadWriteCloser
}

func (c *ConnWrap) Read(p []byte) (n int, err error) {
	n, err = c.conn.Read(p)
	//  log.Printf("Read[%d](%v): %s",n,err,string(p))
	return
}

func (c *ConnWrap) Write(p []byte) (n int, err error) {
	n, err = c.conn.Write(p)
	log.Printf("Write[%d](%v): %s", n, err, string(p))
	return
}

func (c *ConnWrap) Close() (err error) {
	err = c.conn.Close()
	log.Printf("Close(%v)", err)
	return
}

func newBaseFeed(sp SessionProvider, callback Callback, feedType FeedType) (feed *baseFeed, err error) {
	feed = &baseFeed{quit: make(chan interface{}), feedType: feedType, callback: callback}
	go feed.mainLoop(feed.quit, sp)
	return
}

func (f *baseFeed) mainLoop(quit chan interface{}, sp SessionProvider) {
	defer func() { f.conn.Close() }() // In func, since conn will be changed
	go func() {
		connectDelay := makeDelayRetry("Sleeping %dms before next reconnect")
		for quit != nil {
			var conn *tls.Conn
			key, url, err := sp()
			if err == nil {
				conn, err = tls.Dial("tcp", url, DefaultTLS)
			}
			if err != nil {
				log.Printf("Error connection: %+v", err)
			} else {
				connw := &ConnWrap{conn}
				f.conn = connw
				enc := json.NewEncoder(connw) // Dont assign befor our login, let other writers fail on old connection
				// Login
				enc.Encode(&FeedCmd{Cmd: "login", Args: &loginArgs{SessionKey: key}})
				f.encoder = enc
				f.decoder = json.NewDecoder(connw)
				f.decoder.UseNumber()

				f.callback.OnConnect(f.Write, f.feedType)

				var readerr error
				for readerr == nil && quit != nil {
					msg := &FeedMsg{}
					if readerr = f.decoder.Decode(msg); readerr == nil {
						f.callback.OnMessage(msg, f.feedType)
					} else {
						f.callback.OnError(readerr, f.feedType)
					}
				}
				log.Printf("Got error while reading %+v  (Quit: %+v)", err, quit)
			}
			connectDelay()
		}
	}()
	<-quit
	quit = nil
}

func (f *baseFeed) Write(any *FeedCmd) (err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
			f.callback.OnError(e, f.feedType)
		}
		log.Printf("Writing %+v -> %+v", any, err)
	}()

	err = f.encoder.Encode(any)
	return
}

func (f *baseFeed) SendCmd(cmd string, args interface{}) error {
	return f.Write(&FeedCmd{Cmd: cmd, Args: args})
}

// baseFeed implements the Closer interface
// closes the underlying conneciton
func (f *baseFeed) Close() (err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()
	close(f.quit)
	return
}
