package feed

import (
	"bytes"
	"encoding/json"
	"github.com/Forau/yanngo/remote"
	"log"
)

// Convenience function to recive feed messages
type FeedClient func(msg *FeedMsg)

// Implement SubHandler
func (fc FeedClient) Handle(topic string, data []byte) (err error) {
	der := json.NewDecoder(bytes.NewReader(data))
	der.UseNumber()
	var msg FeedMsg
	if err = der.Decode(&msg); err == nil {
		fc(&msg)
	} else {
		log.Printf("WARNING: %v", err)
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
