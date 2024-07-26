package util

import (
	"fmt"
	"time"
)

// const DateFormat_Date = "2006-01-02"

var (
	MinDateTime = time.Unix(-2208988800, 0) // Jan 1, 1900
	MaxDateTime = MinDateTime.Add(1<<63 - 1)

	DefaultTimezone = time.FixedZone("GMT+7", 7)
)

// type CustomTime_DateOnly struct {
// 	time.Time
// }

// func (ct *CustomTime_DateOnly) UnmarshalJSON(b []byte) (err error) {
// 	s := strings.Trim(string(b), "\"")
// 	if s == "null" {
// 		ct.Time = time.Time{}
// 		return
// 	}
// 	ct.Time, err = time.Parse(DateFormat_Date, s)
// 	return
// }

type TimeSpec struct {
	StartDatetime time.Time
	EndDatetime   time.Time
}

var (
	default_StartDatetime = MinDateTime
	default_EndDatetime   = MaxDateTime
)

func (s *TimeSpec) SetDefaultForZeroValues() {
	if s.StartDatetime.IsZero() {
		mainLog.Warn("Found zero value for StartDatetime, reverting to default value (minTime)")
		s.StartDatetime = default_StartDatetime
	}
	if s.EndDatetime.IsZero() {
		mainLog.Warn("Found zero value for EndDatetime, reverting to default value (maxTime)")
		s.EndDatetime = default_EndDatetime
	}
}

func (s TimeSpec) ValidateZeroValues() error {
	if s.StartDatetime.IsZero() || s.EndDatetime.IsZero() {
		return fmt.Errorf("this TimeSpec must have both StartDateTime and EndDateTime to be greater than zero")
	}
	return nil
}
