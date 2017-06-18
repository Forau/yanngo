// A small helper to parse time to nasdaq Stockholm time, and guess open times
package omxtime

import (
	"time"
)

const (
	dteFormat  = "2006-01-02"
	fullFormat = "2006-01-02 15:04:05 MST"

	MinuteX1  = 60000
	MinuteX3  = MinuteX1 * 3
	MinuteX5  = MinuteX1 * 5
	MinuteX10 = MinuteX1 * 10
	MinuteX15 = MinuteX1 * 15
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

func MillisToString(millis int64) string {
	return time.Unix(millis/1000, millis%1000).In(omxloc).Format(fullFormat)
}

func MillisToDayString(millis int64) string {
	return time.Unix(millis/1000, millis%1000).In(omxloc).Format(dteFormat)
}

func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse(dteFormat, dateStr)
}

type OmxTime struct {
	Date      string
	Millis    int64
	OmxOpen   int64
	OmxClose  int64
	DayOfWeek int
}

func NewOmxTimeNow() *OmxTime {
	dte := time.Now().In(omxloc)
	return (&OmxTime{}).Init(dte)
}
func NewOmxTimeMillis(millis int64) *OmxTime {
	dte := time.Unix(millis/1000, millis%1000).In(omxloc)
	return (&OmxTime{}).Init(dte)
}

func NewOmxTimeDate(dateStr string) (*OmxTime, error) {
	if dte, err := ParseDate(dateStr); err != nil {
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
	// When the day starts....
	ot.Millis = time.Date(y, m, d, 0, 0, 0, 0, omxloc).Unix() * 1000

	ot.DayOfWeek = int(dte.Weekday())
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

func (ot *OmxTime) IsTrading(timeMillis int64) bool {
	return timeMillis >= ot.OmxOpen && timeMillis < ot.OmxClose
}

func (ot *OmxTime) TimeSlot(timeMillis int64, scaleMillis int64) (idx int64) {
	if ot.IsTrading(timeMillis) {
		return (timeMillis - ot.OmxOpen) / scaleMillis
	} else {
		return -1
	}
}

/*
// A bit expencive if outside trading hours. Check IsTrading first
func (ot *OmxTime) IsToday(millis int64) bool {
	return ot.IsTrading(millis) || time.Unix(millis/1000, millis%1000).In(omxloc).Format(dteFormat) == ot.Date
}
*/

func (ot *OmxTime) IsToday(millis int64) bool {
	return millis >= ot.Millis && millis < ot.Millis+1000*60*60*24
}
