// A small helper to parse time to nasdaq Stockholm time, and guess open times
package omxtime

import (
	"time"
)

const (
	dteFormat = "2006-01-02"
)

var (
	omxloc *time.Location
)

func init() {
	if loc, err := time.LoadLocation("Europe/Stockholm"); err != nil {
		panic(err)
	} else {
		omxloc = loc
	}
}

type OmxTime struct {
	Date     string
	OmxOpen  int64
	OmxClose int64
}

func NewOmxTimeMillis(millis int64) *OmxTime {
	dte := time.Unix(millis/1000, millis%1000).In(omxloc)
	return (&OmxTime{}).Init(dte)
}

func NewOmxTimeDate(dateStr string) (*OmxTime, error) {
	if dte, err := time.Parse(dteFormat, dateStr); err != nil {
		return nil, err
	} else {
		return (&OmxTime{}).Init(dte.In(omxloc)), nil
	}
}

func (ot *OmxTime) Init(dte time.Time) *OmxTime {
	ot.Date = dte.Format(dteFormat)
	y, m, d := dte.Year(), dte.Month(), dte.Day()

	if dte.Weekday()%6 > 0 { // 0 and 6 is sunday and saterday
		ot.OmxOpen = time.Date(y, m, d, 9, 0, 0, 0, omxloc).Unix() * 1000
		ot.OmxClose = time.Date(y, m, d, 17, 30, 0, 0, omxloc).Unix() * 1000
	} else {
		ot.OmxOpen, ot.OmxClose = -1, -1
	}
	return ot
}

func findTradingDay(dte time.Time, dayCount int) *OmxTime {
	for dte = dte.AddDate(0, 0, dayCount); dte.Weekday()%6 == 0; dte = dte.AddDate(0, 0, dayCount) {
	}
	return (&OmxTime{}).Init(dte)
}

func (ot *OmxTime) PrevTradingDay() *OmxTime {
	dte, _ := time.Parse(dteFormat, ot.Date)
	return findTradingDay(dte.In(omxloc), -1)
}

func (ot *OmxTime) NextTradingDay() *OmxTime {
	dte, _ := time.Parse(dteFormat, ot.Date)
	return findTradingDay(dte.In(omxloc), 1)
}
