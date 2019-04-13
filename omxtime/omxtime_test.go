package omxtime_test

import (
	"github.com/Forau/yanngo/omxtime"
	"testing"
)

func TestOmxTradingInitWithMillis(t *testing.T) {
	input := []struct {
		Millis  int64
		Trading bool
	}{
		{1475557620000, false},
		{1474903740000, true},
		{1474904100000, false},
	}

	for _, d := range input {
		ot := omxtime.NewOmxTimeMillis(d.Millis)
		t.Logf("OmxTime: %+v", ot)
		trading := ot.OmxOpen <= d.Millis && ot.OmxClose > d.Millis
		if trading != d.Trading {
			t.Errorf("Expected trading to be %v, but was %v", d.Trading, trading)
		} else {
			t.Logf("Market open: %v", trading)
		}
	}
}

func TestOmxTradingInitWithString(t *testing.T) {
	input := []struct {
		Date               string
		Error              bool
		PrevDate, NextDate string
	}{
		{"2016-01-01", false, "2015-12-31", "2016-01-04"},
		{"2017-33-12", true, "", ""},
		{"2013-01-22", false, "2013-01-21", "2013-01-23"},
		{"2015-10-10", false, "2015-10-09", "2015-10-12"},
	}

	for _, d := range input {
		ot, err := omxtime.NewOmxTimeDate(d.Date)
		t.Logf("OmxTime: %+v - %+v", ot, err)
		if d.Error && err != nil {
			t.Logf("Got error as expected: %+v", err)
		} else if d.Error && err == nil {
			t.Errorf("Expected error, but didnt get any.  %v, %v\n", d.Error, err)
		} else if !d.Error && err != nil {
			t.Errorf("Got unexpected error: %+v", err)
		} else {
			otPrev := ot.PrevTradingDay()
			t.Logf("PREV: %+v", otPrev)
			if otPrev.Date != d.PrevDate {
				t.Errorf("expected prev day to be %s, but was %s", d.PrevDate, otPrev.Date)
			}
			otNext := ot.NextTradingDay()
			t.Logf("NEXT: %+v", otNext)
			if otNext.Date != d.NextDate {
				t.Errorf("expected next day to be %s, but was %s", d.NextDate, otNext.Date)
			}
		}
	}
}
