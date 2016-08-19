package tinypubsub_test

import (
	"github.com/Forau/yanngo/remote"
	"github.com/Forau/yanngo/remote/tinypubsub"

	"testing"
	"time"
)

func TestSubscribeAndRecive(t *testing.T) {
	doneChan := make(chan string)
	ps := tinypubsub.NewTinyPubSub()
	defer ps.Close()

	ps.Sub("testchan", remote.SubHandlerFn(func(topic string, data []byte) error {
		t.Logf("Got msg on topic %v: %v", topic, string(data))
		doneChan <- string(data)
		return nil
	}))

	ps.Pub("theWrongChan", []byte("Nobody Listening"))
	dataToSend := []string{"Hello! Test is here", "Second message!"}
	for _, dta := range dataToSend {
		ps.Pub("testchan", []byte(dta))
	}

	for _, exp := range dataToSend {
		select {
		case res := <-doneChan:
			if res != exp {
				t.Errorf("Expected %v, but got %v", exp, res)
				return
			}
		case <-time.After(1 * time.Second):
			t.Errorf("Did not succeed within 1 second. Was expecting: %v", exp)
			return
		}
	}
}
