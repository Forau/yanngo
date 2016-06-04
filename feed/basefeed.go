package feed

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/Forau/yanngo/swagger"
	"io"
	"log"
)

// Used when sending feed commands
type FeedCmd struct {
	Cmd  string      `json:"cmd"`
	Args interface{} `json:"args"`
}

// Arguments for getting orders and trades when logging in
type getState struct {
	DeletedOrders bool  `json:"deleted_orders"`
	Days          int64 `json:"days,omitempty"`
}

// Arguments for sending the login command
type loginArgs struct {
	SessionKey string      `json:"session_key"`
	GetState   interface{} `json:"get_state,omitempty"`
}

type baseFeed struct {
	conn    io.ReadWriteCloser
	encoder *json.Encoder
	decoder *json.Decoder
}

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

func newBaseFeed(state *swagger.Login, feedData swagger.Feed) (feed *baseFeed, err error) {
	log.Printf("Connecting to: %s", fmt.Sprintf("%s:%d", feedData.Hostname, feedData.Port))
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", feedData.Hostname, feedData.Port), nil)
	if err != nil {
		return nil, err
	}
	connw := &ConnWrap{conn}
	log.Printf("Connected: %+v", connw)
	feed = &baseFeed{connw, json.NewEncoder(connw), json.NewDecoder(connw)}
	err = feed.login(state.SessionKey)
	return
}

// baseFeed implements the Writer interface
func (f *baseFeed) Write(any interface{}) error {
	return f.encoder.Encode(any)
}

func (f *baseFeed) SendCmd(cmd string, args interface{}) error {
	return f.Write(&FeedCmd{Cmd: cmd, Args: args})
}

// baseFeed implements the Closer interface
// closes the underlying conneciton
func (f *baseFeed) Close() error {
	return f.conn.Close()
}

// Send the login command with the specified session key
func (f *baseFeed) login(session string) error {
	getState := &getState{DeletedOrders: true} // TODO: Only send on private?
	return f.SendCmd("login", &loginArgs{SessionKey: session, GetState: getState})
}
