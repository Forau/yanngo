package feed

import (
	"crypto/tls"
	"encoding/json"
	"github.com/Forau/yanngo/feed/feedmodel"
	"io"
	"log"
	"time"
)

// Set this static variable if you need special TLS config.  Used in tests for example.
var DefaultTLS *tls.Config

type CmdWriter func(cmd *feedmodel.FeedCmd) error

type Callback interface {
	OnConnect(w CmdWriter, ft feedmodel.FeedType)
	OnMessage(msg *feedmodel.FeedMsg, ft feedmodel.FeedType)
	OnError(err error, ft feedmodel.FeedType)
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

	feedType feedmodel.FeedType
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

func newBaseFeed(sp SessionProvider, callback Callback, feedType feedmodel.FeedType) (feed *baseFeed, err error) {
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
				enc.Encode(&feedmodel.FeedCmd{Cmd: "login", Args: &loginArgs{SessionKey: key}})
				f.encoder = enc
				f.decoder = json.NewDecoder(connw)
				f.decoder.UseNumber()

				f.callback.OnConnect(f.Write, f.feedType)

				var readerr error
				for readerr == nil && quit != nil {
					msg := &feedmodel.FeedMsg{}
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

func (f *baseFeed) Write(any *feedmodel.FeedCmd) (err error) {
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
	return f.Write(&feedmodel.FeedCmd{Cmd: cmd, Args: args})
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
