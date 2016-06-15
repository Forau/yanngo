package feed

import (
	"encoding/json"
	"fmt"
	"github.com/Forau/yanngo/api"
	"log"
)

type callback struct {
	subs []*FeedCmd
}

func (c *callback) OnConnect(w CmdWriter, ft FeedType) {
	log.Printf("Connect[%v]: %+v", ft, w)
	if ft == PublicFeedType {
		for _, s := range c.subs {
			log.Printf("Sending: %+v", s)
			w(s)
		}
	}
}
func (c *callback) OnMessage(msg *FeedMsg, ft FeedType) { log.Printf("Msg[%v]: %+v", ft, msg.String()) }
func (c *callback) OnError(err error, ft FeedType)      { log.Printf("Err[%v]: %+v", ft, err) }

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

func MakePrivateSessionProvider(api *api.ApiClient) SessionProvider {
	return func() (key, url string, err error) {
		session, err := api.Session() // TODO: Error check
		if err != nil {
			return "", "", err
		}
		return session.SessionKey, fmt.Sprintf("%s:%d", session.PrivateFeed.Hostname, session.PrivateFeed.Port), nil
	}
}
func MakePublicSessionProvider(api *api.ApiClient) SessionProvider {
	return func() (key, url string, err error) {
		session, err := api.Session() // TODO: Error check
		if err != nil {
			return "", "", err
		}
		return session.SessionKey, fmt.Sprintf("%s:%d", session.PublicFeed.Hostname, session.PublicFeed.Port), nil
	}
}

// This object should connect to the feeds, keep them alive, and keep state
type FeedDaemon struct {
	private *baseFeed
	public  *baseFeed
}

// This is not the preferred constuctor.
func NewFeedDaemonAPI(api *api.ApiClient) (fd *FeedDaemon, err error) {
	// Dummy callback, that just subscribes to ERIC price every reconnect.
	cb := &callback{
		subs: []*FeedCmd{
			&FeedCmd{Cmd: "subscribe", Args: map[string]interface{}{"t": "price", "i": "101", "m": 11}},
		},
	}

	return NewFeedDaemon(MakePrivateSessionProvider(api), MakePublicSessionProvider(api), cb)
}

func NewFeedDaemon(privSess, pubSess SessionProvider, cb Callback) (fd *FeedDaemon, err error) {
	fd = &FeedDaemon{}

	fd.private, err = newBaseFeed(privSess, cb, PrivateFeedType)
	if err != nil {
		return
	}
	fd.public, err = newBaseFeed(pubSess, cb, PublicFeedType)
	return
}

func (fd *FeedDaemon) Close() error {
	err1 := fd.private.Close()
	err2 := fd.public.Close()
	if err1 != nil {
		return err1
	}
	return err2
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
	cmd := &FeedCmd{
		Cmd:  "subscribe",
		Args: args,
	}
	return fd.public.Write(cmd)
}

// ONLY for testing....
func (fd *FeedDaemon) KillSockets() (string, error) {
	err := fd.private.conn.Close()
	if err != nil {
		return "", err
	}
	err = fd.public.conn.Close()
	return "Killed sockets", err
}
