package models

import (
	"strings"
	"time"
)

const timeFormat = "2006-01-02"

type PaymentExtractionSchedule struct {
	Date   JsonDate `json:"date" example:"2020-01-01"`
	Amount float32  `json:"amount" example:"100.21"`
}

type JsonDate time.Time

var (
	timeZone *time.Location
)

func init() {
	var err error
	timeZone, err = time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
}

func (j *JsonDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	t, err := time.ParseInLocation(timeFormat, s, timeZone)
	if err != nil {
		return err
	}
	*j = JsonDate(t)
	return nil
}

func (j JsonDate) MarshalJSON() ([]byte, error) {
	t := time.Time(j)
	formatted := t.Format(timeFormat)
	jsonStr := "\"" + formatted + "\""
	return []byte(jsonStr), nil
}
