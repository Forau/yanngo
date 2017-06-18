// A small utillity for checking timestamps for omx. Could be better without flags, but its quick and dirty...
package main

import (
	"github.com/Forau/yanngo/omxtime"

	"flag"
	"fmt"
	"time"
)

var (
	date = flag.String("date", "", "Date to start at, in format: 2006-01-02")
	ms   = flag.Int64("ms", 0, "Date to start at, in millis")
)

func printOmxTime(ot *omxtime.OmxTime) {
	fmt.Printf("%s: ", ot.Date)
	if ot.OmxOpen > 0 {
		fmt.Printf("Open from %d to %d. ", ot.OmxOpen, ot.OmxClose)
	} else {
		fmt.Printf("Closed. ")
	}
	fmt.Printf("Previous: %s, Next: %s\n", ot.PrevTradingDay().Date, ot.NextTradingDay().Date)
}

func main() {
	var ot *omxtime.OmxTime
	var err error

	var millis int64

	flag.Parse()

	if *date != "" {
		if ot, err = omxtime.NewOmxTimeDate(*date); err != nil {
			panic(err)
		}
	} else {
		if *ms > 0 {
			millis = *ms
		} else {
			millis = time.Now().Unix() * 1000
		}
		ot = omxtime.NewOmxTimeMillis(millis)
	}
	printOmxTime(ot)

	if millis > 0 {
		fmt.Printf("%s - 5min candle nr: %d\n", omxtime.MillisToString(millis), ot.TimeSlot(millis, 5*60000))
	}
}
