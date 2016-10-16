package feed_test

import (
	"fmt"
	"github.com/Forau/yanngo/feed"
	"github.com/Forau/yanngo/feed/feedmodel"
	"github.com/Forau/yanngo/remote/tinypubsub"
	"testing"
	"time"
)

func TestFeedClient(t *testing.T) {
	pubsub := tinypubsub.NewTinyPubSub()

	arrToSend := []struct {
		Type, Data string
		expected   bool
	}{
		{"whatever", `"And we dont care"`, true},
		{"somethingElse", `{"is":"this","json":"?"}`, true},
		{"parseError", `"is":"this","json":!}`, false},
		{"news", `{"news_id":553027001,"source_id":1,
              "headline":"RÅVAROR: STARKA RÅVARUAVANCEMANG EFTER SÄMRE JOBBRAPPORT",
              "lang":"sv","type":"Market commentary","timestamp":1472842201000}`, true},
	}

	msgChan := make(chan *feedmodel.FeedMsg, 1)

	feed.FeedClient(func(msg *feedmodel.FeedMsg) {
		t.Log("Msg: ", msg)
		msgChan <- msg
	}).Bind(pubsub, "test.feed")

	for _, ts := range arrToSend {
		pubsub.Pub("test.feed", []byte(fmt.Sprintf(`{"type":"%s", "data":%v}`, ts.Type, ts.Data)))
		select {
		case res := <-msgChan:
			if !ts.expected {
				t.Errorf("Got %+v, but was not expected", res)
			} else if res.Type != ts.Type || string(res.Data) != ts.Data {
				t.Errorf("Expected %+v, but got %+v", ts, res)
			}
		case <-time.After(10 * time.Millisecond):
			if ts.expected {
				t.Errorf("Timeout waiting for: %+v", ts)
			} else {
				t.Logf("Got timeout on %+v as expected", ts)
			}
		}
	}
	close(msgChan)
}
