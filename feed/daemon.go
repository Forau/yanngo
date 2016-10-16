package feed

import (
	"fmt"
	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/feed/feedmodel"
	//	"log"
)

// A representation of the subscription. Can generate the proper types depending on data
type FeedSubscriptionKey struct {
	T, I, M string
	S       int64
	Delay   bool
}

func (fsk *FeedSubscriptionKey) ToFeedCmd(cmdType string) (ret *feedmodel.FeedCmd, err error) {
	ret = &feedmodel.FeedCmd{Cmd: cmdType}
	switch fsk.T {
	case "price", "depth", "trade", "trading_status":
		args := &feedCmdArgs{T: fsk.T, I: fsk.I}
		_, err = fmt.Sscan(fsk.M, &args.M)
		ret.Args = args
	case "indicator":
		ret.Args = &indicatorArgs{T: fsk.T, I: fsk.I, M: fsk.M}
	case "news":
		ret.Args = &newsArgs{T: fsk.T, S: fsk.S, Delay: fsk.Delay}
	}
	return
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
	cb := &FeedState{
		subs: []*feedmodel.FeedCmd{
			&feedmodel.FeedCmd{Cmd: "subscribe", Args: map[string]interface{}{"t": "price", "i": "101", "m": 11}},
		},
	}

	return NewFeedDaemon(MakePrivateSessionProvider(api), MakePublicSessionProvider(api), cb)
}

func NewFeedDaemon(privSess, pubSess SessionProvider, cb Callback) (fd *FeedDaemon, err error) {
	fd = &FeedDaemon{}

	fd.private, err = newBaseFeed(privSess, cb, feedmodel.PrivateFeedType)
	if err != nil {
		return
	}
	fd.public, err = newBaseFeed(pubSess, cb, feedmodel.PublicFeedType)
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

func (fd *FeedDaemon) Subscribe(typ, ident, market string) error {
	cmd, e := (&FeedSubscriptionKey{T: typ, I: ident, M: market}).ToFeedCmd("subscribe")
	if e != nil {
		return e
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
